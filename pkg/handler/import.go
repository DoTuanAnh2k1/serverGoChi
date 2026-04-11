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

// HandlerImport handles POST /aa/import
// Accepts import.txt content as plain text body.
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
				Name:        cols[0],
				SiteName:    cols[1],
				IPAddress:   cols[2],
				Port:        int32(port),
				Namespace:   cols[4],
				Description: cols[5],
				SystemType:  "5GC",
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
			neMap[ne.Name] = ne.ID
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

	logger.Logger.WithField("actor", actor.Username).Infof("import: %d results", len(results))

	saveHistory(db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("import (%d items)", len(results)),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
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
