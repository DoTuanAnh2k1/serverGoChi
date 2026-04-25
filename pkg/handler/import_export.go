package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/token"
)

// ── Export CSV ──────────────────────────────────────────────────────────

// exportAuth checks JWT from either Authorization header or _token query param
// (for direct link downloads where the browser can't set headers).
func exportAuth(w http.ResponseWriter, r *http.Request) bool {
	tok := r.URL.Query().Get("_token")
	if tok == "" {
		tok = r.Header.Get("Authorization")
	}
	if tok == "" {
		response.Unauthorized(w)
		return false
	}
	_, _, err := token.ParseToken(tok)
	if err != nil {
		response.Unauthorized(w)
		return false
	}
	return true
}

func HandlerExportUsers(w http.ResponseWriter, r *http.Request) {
	if !exportAuth(w, r) { return }
	users, err := service.ListUsers()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="users_%s.csv"`, time.Now().Format("20060102_150405")))
	cw := csv.NewWriter(w)
	cw.Write([]string{"username", "email", "full_name", "phone", "role", "is_enabled"})
	for _, u := range users {
		cw.Write([]string{
			u.Username, u.Email, u.FullName, u.Phone, u.Role,
			strconv.FormatBool(u.IsEnabled),
		})
	}
	cw.Flush()
}

func HandlerExportNEs(w http.ResponseWriter, r *http.Request) {
	if !exportAuth(w, r) { return }
	nes, err := service.ListNEs()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="nes_%s.csv"`, time.Now().Format("20060102_150405")))
	cw := csv.NewWriter(w)
	cw.Write([]string{"namespace", "ne_type", "site_name", "description", "master_ip", "master_port", "ssh_username", "ssh_password", "command_url", "conf_mode"})
	for _, n := range nes {
		cw.Write([]string{
			n.Namespace, n.NeType, n.SiteName, n.Description,
			n.MasterIP, strconv.Itoa(int(n.MasterPort)),
			n.SSHUsername, "", // never export ssh_password
			n.CommandURL, n.ConfMode,
		})
	}
	cw.Flush()
}

func HandlerExportCommands(w http.ResponseWriter, r *http.Request) {
	if !exportAuth(w, r) { return }
	cmds, err := service.ListCommands(0, "")
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	// Resolve NE namespaces for readability.
	nes, _ := service.ListNEs()
	neMap := make(map[int64]string, len(nes))
	for _, n := range nes {
		neMap[n.ID] = n.Namespace
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="commands_%s.csv"`, time.Now().Format("20060102_150405")))
	cw := csv.NewWriter(w)
	cw.Write([]string{"ne_namespace", "service", "cmd_text", "description"})
	for _, c := range cmds {
		cw.Write([]string{
			neMap[c.NeID], c.Service, c.CmdText, c.Description,
		})
	}
	cw.Flush()
}

// ── Import CSV ──────────────────────────────────────────────────────────

func HandlerImportUsers(w http.ResponseWriter, r *http.Request) {
	records, err := parseCSVFromRequest(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(records) < 2 {
		response.Write(w, http.StatusBadRequest, "CSV must have a header row and at least one data row")
		return
	}
	header := normalizeHeader(records[0])
	created, skipped, errors := 0, 0, []string{}
	for i, row := range records[1:] {
		m := zipRow(header, row)
		username := m["username"]
		password := m["password"]
		if username == "" {
			errors = append(errors, fmt.Sprintf("row %d: missing username", i+2))
			continue
		}
		if password == "" {
			password = "Changeme1!" // default if not provided
		}
		role := m["role"]
		if role == "" {
			role = db_models.RoleUser
		}
		enabled := true
		if v := m["is_enabled"]; v == "false" || v == "0" {
			enabled = false
		}
		u := &db_models.User{
			Username:  username,
			Email:     m["email"],
			FullName:  m["full_name"],
			Phone:     m["phone"],
			Role:      role,
			IsEnabled: enabled,
		}
		if err := service.CreateUser(u, password); err != nil {
			if strings.Contains(err.Error(), "already taken") {
				skipped++
			} else {
				errors = append(errors, fmt.Sprintf("row %d (%s): %v", i+2, username, err))
			}
			continue
		}
		created++
	}
	response.Write(w, http.StatusOK, map[string]interface{}{
		"created": created, "skipped": skipped, "errors": errors,
	})
}

func HandlerImportNEs(w http.ResponseWriter, r *http.Request) {
	records, err := parseCSVFromRequest(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(records) < 2 {
		response.Write(w, http.StatusBadRequest, "CSV must have a header row and at least one data row")
		return
	}
	header := normalizeHeader(records[0])
	created, skipped, errors := 0, 0, []string{}
	for i, row := range records[1:] {
		m := zipRow(header, row)
		ns := m["namespace"]
		if ns == "" {
			errors = append(errors, fmt.Sprintf("row %d: missing namespace", i+2))
			continue
		}
		port, _ := strconv.Atoi(m["master_port"])
		n := &db_models.NE{
			Namespace:   ns,
			NeType:      m["ne_type"],
			SiteName:    m["site_name"],
			Description: m["description"],
			MasterIP:    m["master_ip"],
			MasterPort:  int32(port),
			SSHUsername:  m["ssh_username"],
			SSHPassword:  m["ssh_password"],
			CommandURL:  m["command_url"],
			ConfMode:    m["conf_mode"],
		}
		if err := service.CreateNE(n); err != nil {
			if strings.Contains(err.Error(), "already") {
				skipped++
			} else {
				errors = append(errors, fmt.Sprintf("row %d (%s): %v", i+2, ns, err))
			}
			continue
		}
		created++
	}
	response.Write(w, http.StatusOK, map[string]interface{}{
		"created": created, "skipped": skipped, "errors": errors,
	})
}

func HandlerImportCommands(w http.ResponseWriter, r *http.Request) {
	records, err := parseCSVFromRequest(r)
	if err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	if len(records) < 2 {
		response.Write(w, http.StatusBadRequest, "CSV must have a header row and at least one data row")
		return
	}
	// Build NE namespace→ID map.
	nes, _ := service.ListNEs()
	neMap := make(map[string]int64, len(nes))
	for _, n := range nes {
		neMap[strings.ToLower(n.Namespace)] = n.ID
	}
	header := normalizeHeader(records[0])
	created, skipped, errors := 0, 0, []string{}
	for i, row := range records[1:] {
		m := zipRow(header, row)
		ns := m["ne_namespace"]
		neID, ok := neMap[strings.ToLower(ns)]
		if !ok {
			// Try ne_id directly.
			if id, err := strconv.ParseInt(m["ne_id"], 10, 64); err == nil && id > 0 {
				neID = id
			} else {
				errors = append(errors, fmt.Sprintf("row %d: unknown NE namespace %q", i+2, ns))
				continue
			}
		}
		svc := m["service"]
		cmdText := m["cmd_text"]
		if cmdText == "" {
			errors = append(errors, fmt.Sprintf("row %d: missing cmd_text", i+2))
			continue
		}
		c := &db_models.Command{
			NeID:        neID,
			Service:     svc,
			CmdText:     cmdText,
			Description: m["description"],
		}
		if err := service.CreateCommand(c); err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already") {
				skipped++
			} else {
				errors = append(errors, fmt.Sprintf("row %d: %v", i+2, err))
			}
			continue
		}
		created++
	}
	response.Write(w, http.StatusOK, map[string]interface{}{
		"created": created, "skipped": skipped, "errors": errors,
	})
}

// ��─ CSV helpers ─────────────────────────────────────────────────────────

func parseCSVFromRequest(r *http.Request) ([][]string, error) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return nil, fmt.Errorf("parse form: %w", err)
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("missing 'file' field: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	// Remove BOM if present.
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	reader := csv.NewReader(bytes.NewReader(data))
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	return reader.ReadAll()
}

func normalizeHeader(row []string) []string {
	out := make([]string, len(row))
	for i, h := range row {
		out[i] = strings.ToLower(strings.TrimSpace(h))
	}
	return out
}

func zipRow(header, row []string) map[string]string {
	m := make(map[string]string, len(header))
	for i, h := range header {
		if i < len(row) {
			m[h] = strings.TrimSpace(row[i])
		}
	}
	return m
}
