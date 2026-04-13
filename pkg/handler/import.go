package handler

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/bcrypt"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

// HandlerImport nhập hàng loạt dữ liệu cấu hình từ file text có cấu trúc section.
//
// Input : POST body plain text theo định dạng import.txt:
//         [users]       username,password
//         [nes]         name,site_name,ip_address,port,namespace,description
//         [roles]       permission,scope,ne_type,include_type,path
//         [user_roles]  username,permission
//         [user_nes]    username,ne_name
//         (dòng đầu mỗi section là header — bị bỏ qua; dòng bắt đầu # là comment)
// Output: 200 [ { type, name, status, detail } ] — kết quả từng item (ok/skip/error)
//         500 nếu không lấy được actor từ context
// Flow  : lấy actor từ context → parseSections(body) → xử lý tuần tự từng section
//         (users → nes → roles → user_roles → user_nes) → ghi operation history →
//         trả mảng kết quả
func HandlerImport(w http.ResponseWriter, r *http.Request) {
	actor, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		response.InternalError(w, "Internal Server Error")
		return
	}

	sections := parseSections(r.Body)
	db := store.GetSingleton()
	var results []importResult

	// Users
	if rows, ok := sections["users"]; ok {
		for _, cols := range rows {
			if len(cols) < 2 {
				continue
			}
			username, password := cols[0], cols[1]
			existing, _ := db.GetUserByUserName(username)
			if existing != nil {
				results = append(results, importResult{"user", username, "skip", "already exists"})
				continue
			}
			now := time.Now()
			user := &db_models.TblAccount{
				AccountName:    username,
				Password:       bcrypt.Encode(username + password),
				AccountType:    2,
				IsEnable:       true,
				Status:         true,
				CreatedDate:    now,
				UpdatedDate:    now,
				LastLoginTime:  now,
				LastChangePass: now,
				LockedTime:     now,
				CreatedBy:      actor.Username,
			}
			if err := db.AddUser(user); err != nil {
				results = append(results, importResult{"user", username, "error", err.Error()})
			} else {
				results = append(results, importResult{"user", username, "ok", "created"})
			}
		}
	}

	// NEs
	if rows, ok := sections["nes"]; ok {
		for _, cols := range rows {
			if len(cols) < 6 {
				continue
			}
			port, _ := strconv.Atoi(cols[3])
			ne := &db_models.CliNe{
				NeName:            cols[0],
				SiteName:          cols[1],
				ConfMasterIP:      cols[2],
				ConfPortMasterSSH: int32(port),
				Namespace:         cols[4],
				Description:       cols[5],
				SystemType:        "5GC",
			}
			if err := db.CreateCliNe(ne); err != nil {
				results = append(results, importResult{"ne", cols[0], "error", err.Error()})
			} else {
				results = append(results, importResult{"ne", cols[0], "ok", fmt.Sprintf("created (id=%d)", ne.ID)})
			}
		}
	}

	// Roles
	if rows, ok := sections["roles"]; ok {
		for _, cols := range rows {
			if len(cols) < 5 {
				continue
			}
			role := &db_models.CliRole{
				Permission:  cols[0],
				Scope:       cols[1],
				NeType:      cols[2],
				IncludeType: cols[3],
				Path:        cols[4],
			}
			existing, _ := db.GetCliRole(role)
			if existing != nil {
				results = append(results, importResult{"role", cols[0], "skip", "already exists"})
				continue
			}
			if err := db.CreateCliRole(role); err != nil {
				results = append(results, importResult{"role", cols[0], "error", err.Error()})
			} else {
				results = append(results, importResult{"role", cols[0], "ok", "created"})
			}
		}
	}

	// User-role mappings
	if rows, ok := sections["user_roles"]; ok {
		for _, cols := range rows {
			if len(cols) < 2 {
				continue
			}
			username, permission := cols[0], cols[1]
			user, err := db.GetUserByUserName(username)
			if err != nil || user == nil {
				results = append(results, importResult{"user_role", username + " -> " + permission, "error", "user not found"})
				continue
			}
			if err := db.AddRole(&db_models.CliRoleUserMapping{UserID: user.AccountID, Permission: permission}); err != nil {
				results = append(results, importResult{"user_role", username + " -> " + permission, "error", err.Error()})
			} else {
				results = append(results, importResult{"user_role", username + " -> " + permission, "ok", "assigned"})
			}
		}
	}

	// User-NE mappings
	if rows, ok := sections["user_nes"]; ok {
		neMap := map[string]int64{}
		allNes, _ := db.GetCliNeListBySystemType("5GC")
		for _, ne := range allNes {
			neMap[ne.NeName] = ne.ID
		}
		for _, cols := range rows {
			if len(cols) < 2 {
				continue
			}
			username, neName := cols[0], cols[1]
			user, err := db.GetUserByUserName(username)
			if err != nil || user == nil {
				results = append(results, importResult{"user_ne", username + " -> " + neName, "error", "user not found"})
				continue
			}
			neID, found := neMap[neName]
			if !found {
				results = append(results, importResult{"user_ne", username + " -> " + neName, "error", "ne not found"})
				continue
			}
			if err := db.CreateUserNeMapping(&db_models.CliUserNeMapping{UserID: user.AccountID, TblNeID: neID}); err != nil {
				results = append(results, importResult{"user_ne", username + " -> " + neName, "error", err.Error()})
			} else {
				results = append(results, importResult{"user_ne", username + " -> " + neName, "ok", "assigned"})
			}
		}
	}

	// NE configs (cli_ne_config): ne_name, ip_address, port, username, password, protocol, description
	if rows, ok := sections["ne_configs"]; ok {
		neMap := map[string]int64{}
		allNes, _ := db.GetCliNeListBySystemType("5GC")
		for _, ne := range allNes {
			neMap[ne.NeName] = ne.ID
		}
		for _, cols := range rows {
			if len(cols) < 2 {
				continue
			}
			neName := cols[0]
			neID, found := neMap[neName]
			if !found {
				results = append(results, importResult{"ne_config", neName, "error", "ne not found"})
				continue
			}
			port, _ := strconv.Atoi(safeCol(cols, 2))
			protocol := safeCol(cols, 5)
			if protocol == "" {
				protocol = "SSH"
			}
			cfg := &db_models.CliNeConfig{
				NeID:        neID,
				IPAddress:   safeCol(cols, 1),
				Port:        int32(port),
				Username:    safeCol(cols, 3),
				Password:    safeCol(cols, 4),
				Protocol:    protocol,
				Description: safeCol(cols, 6),
			}
			if err := db.CreateCliNeConfig(cfg); err != nil {
				results = append(results, importResult{"ne_config", neName, "error", err.Error()})
			} else {
				results = append(results, importResult{"ne_config", neName, "ok", fmt.Sprintf("created (id=%d)", cfg.ID)})
			}
		}
	}

	logger.Logger.WithField("actor", actor.Username).Infof("import: %d results", len(results))

	saveHistory(db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("import (%d items)", len(results)),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     actor.Username,
	}, "success")

	response.Write(w, http.StatusOK, results)
}

type importResult struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

// safeCol returns cols[i] if i is within bounds, otherwise empty string.
func safeCol(cols []string, i int) string {
	if i < len(cols) {
		return cols[i]
	}
	return ""
}

// parseSections reads import.txt format from an io.Reader.
func parseSections(r interface{ Read([]byte) (int, error) }) map[string][][]string {
	sections := map[string][][]string{}
	var current string
	headerSkipped := map[string]bool{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			current = strings.ToLower(line[1 : len(line)-1])
			continue
		}
		if current == "" {
			continue
		}
		if !headerSkipped[current] {
			headerSkipped[current] = true
			continue
		}
		cols := strings.Split(line, ",")
		for i := range cols {
			cols[i] = strings.TrimSpace(cols[i])
		}
		sections[current] = append(sections[current], cols)
	}
	return sections
}
