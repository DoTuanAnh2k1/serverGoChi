package sshcli

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// MgtClient talks to cli-mgt-svc over HTTP.
type MgtClient struct {
	BaseURL string
	HTTP    *http.Client
	// Token is the "Basic <jwt>" string returned by /aa/authenticate and
	// expected verbatim as the Authorization header by subsequent routes.
	Token string
	// Role is the "aud" claim from the JWT: "admin" (covers SuperAdmin + Admin)
	// or "user" (Normal).
	Role string
}

// NewMgtClient builds a client with a sane default timeout.
func NewMgtClient(baseURL string) *MgtClient {
	return &MgtClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP:    &http.Client{Timeout: 15 * time.Second},
	}
}

type authResponse struct {
	Status       string `json:"status"`
	ResponseData string `json:"response_data"`
	ResponseCode string `json:"response_code"`
	SystemType   string `json:"system_type"`
}

// Authenticate logs in with username/password and stores the token + role on
// the client. Role is extracted from the JWT's "aud" claim so that the SSH
// server can gate Admin-only access without a second round-trip (and without
// relying on /admin/user/list, which hides SuperAdmin entries).
func (c *MgtClient) Authenticate(username, password string) error {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/aa/authenticate", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("authenticate: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("authenticate failed (status %d): %s", resp.StatusCode, truncate(string(raw), 200))
	}
	var ar authResponse
	if err := json.Unmarshal(raw, &ar); err != nil {
		return fmt.Errorf("decode auth response: %w", err)
	}
	if ar.ResponseData == "" {
		return fmt.Errorf("authenticate: empty token in response")
	}
	c.Token = ar.ResponseData
	role, err := roleFromToken(c.Token)
	if err != nil {
		return fmt.Errorf("extract role: %w", err)
	}
	c.Role = role
	if role != "admin" {
		return fmt.Errorf("role %q is not authorized for the CLI (admin/superadmin required)", role)
	}
	return nil
}

// roleFromToken parses the "aud" claim from the JWT. We trust the token
// because we just received it from mgt-svc, so we decode the payload without
// verifying the signature.
func roleFromToken(tok string) (string, error) {
	tok = strings.TrimPrefix(tok, "Basic ")
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("malformed JWT (want 3 parts, got %d)", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		payload, err = base64.URLEncoding.DecodeString(parts[1])
		if err != nil {
			return "", err
		}
	}
	var claims struct {
		Aud string `json:"aud"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", err
	}
	return claims.Aud, nil
}

// do issues a JSON request with the auth header and returns raw response bytes on 2xx/3xx.
// For list endpoints the server returns 302 Found (http.StatusFound) by convention.
func (c *MgtClient) do(method, path string, body any) ([]byte, int, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.BaseURL+path, reader)
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", c.Token)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return raw, resp.StatusCode, fmt.Errorf("%s %s: status %d: %s", method, path, resp.StatusCode, truncate(string(raw), 300))
	}
	return raw, resp.StatusCode, nil
}

// ---- Users ----

type UserInfo struct {
	AccountID         int64  `json:"account_id"`
	AccountName       string `json:"account_name"`
	FullName          string `json:"full_name"`
	Email             string `json:"email"`
	Address           string `json:"address"`
	PhoneNumber       string `json:"phone_number"`
	AccountType       int    `json:"account_type"`
	Description       string `json:"description"`
	IsEnable          bool   `json:"is_enable"`
	Status            bool   `json:"status"`
	CreatedBy         string `json:"created_by"`
	LoginFailureCount int    `json:"login_failure_count"`
}

func (c *MgtClient) ListUsers() ([]UserInfo, error) {
	raw, _, err := c.do(http.MethodGet, "/aa/admin/user/list", nil)
	if err != nil {
		return nil, err
	}
	var users []UserInfo
	if err := json.Unmarshal(raw, &users); err != nil {
		return nil, fmt.Errorf("decode users: %w (body=%s)", err, truncate(string(raw), 200))
	}
	return users, nil
}

func (c *MgtClient) CreateUser(fields map[string]any) error {
	_, _, err := c.do(http.MethodPost, "/aa/authenticate/user/set", fields)
	return err
}

func (c *MgtClient) UpdateUser(fields map[string]any) error {
	_, _, err := c.do(http.MethodPost, "/aa/admin/user/update", fields)
	return err
}

func (c *MgtClient) DeleteUser(accountName string) error {
	_, _, err := c.do(http.MethodPost, "/aa/authenticate/user/delete", map[string]string{"account_name": accountName})
	return err
}

func (c *MgtClient) ResetUserPassword(username, newPassword string) error {
	_, _, err := c.do(http.MethodPost, "/aa/authenticate/user/reset-password", map[string]string{
		"username":     username,
		"new_password": newPassword,
	})
	return err
}

// ---- NEs ----

type NeInfo struct {
	ID                int64  `json:"id"`
	NeName            string `json:"ne_name"`
	SiteName          string `json:"site_name"`
	Namespace         string `json:"namespace"`
	ConfMasterIP      string `json:"conf_master_ip"`
	ConfPortMasterTCP int    `json:"conf_port_master_tcp"`
	CommandURL        string `json:"command_url"`
	ConfMode          string `json:"conf_mode"`
	Description       string `json:"description"`
	SystemType        string `json:"system_type"`
}

func (c *MgtClient) ListNEs() ([]NeInfo, error) {
	raw, _, err := c.do(http.MethodGet, "/aa/admin/ne/list", nil)
	if err != nil {
		return nil, err
	}
	var nes []NeInfo
	if err := json.Unmarshal(raw, &nes); err != nil {
		return nil, fmt.Errorf("decode nes: %w (body=%s)", err, truncate(string(raw), 200))
	}
	return nes, nil
}

func (c *MgtClient) CreateNE(fields map[string]any) error {
	_, _, err := c.do(http.MethodPost, "/aa/admin/ne/create", fields)
	return err
}

func (c *MgtClient) UpdateNE(fields map[string]any) error {
	_, _, err := c.do(http.MethodPost, "/aa/admin/ne/update", fields)
	return err
}

func (c *MgtClient) DeleteNEByID(id int64) error {
	_, _, err := c.do(http.MethodPost, "/aa/authorize/ne/remove", map[string]int64{"id": id})
	return err
}

// ResolveNEID returns the NE id matching the target, which may be a numeric id
// or an ne_name. Returns an error if not found.
func (c *MgtClient) ResolveNEID(target string) (int64, error) {
	if id, err := strconv.ParseInt(target, 10, 64); err == nil {
		return id, nil
	}
	nes, err := c.ListNEs()
	if err != nil {
		return 0, err
	}
	for _, n := range nes {
		if n.NeName == target {
			return n.ID, nil
		}
	}
	return 0, fmt.Errorf("no NE with name or id %q", target)
}

// User ↔ NE mapping (direct).

func (c *MgtClient) AssignNEToUser(username string, neID int64) error {
	_, _, err := c.do(http.MethodPost, "/aa/authorize/ne/set", map[string]string{
		"username": username,
		"neid":     strconv.FormatInt(neID, 10),
	})
	return err
}

func (c *MgtClient) UnassignNEFromUser(username string, neID int64) error {
	_, _, err := c.do(http.MethodPost, "/aa/authorize/ne/delete", map[string]string{
		"username": username,
		"neid":     strconv.FormatInt(neID, 10),
	})
	return err
}

// ---- Groups ----

type GroupInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *MgtClient) ListGroups() ([]GroupInfo, error) {
	raw, _, err := c.do(http.MethodGet, "/aa/group/list", nil)
	if err != nil {
		return nil, err
	}
	var gs []GroupInfo
	if err := json.Unmarshal(raw, &gs); err != nil {
		return nil, fmt.Errorf("decode groups: %w (body=%s)", err, truncate(string(raw), 200))
	}
	return gs, nil
}

func (c *MgtClient) CreateGroup(fields map[string]any) error {
	_, _, err := c.do(http.MethodPost, "/aa/group/create", fields)
	return err
}

func (c *MgtClient) UpdateGroup(fields map[string]any) error {
	_, _, err := c.do(http.MethodPost, "/aa/group/update", fields)
	return err
}

func (c *MgtClient) DeleteGroupByID(id int64) error {
	_, _, err := c.do(http.MethodPost, "/aa/group/delete", map[string]int64{"id": id})
	return err
}

type GroupDetail struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Users       []string `json:"users"`
	NeIDs       []int64  `json:"ne_ids"`
}

func (c *MgtClient) ShowGroup(id int64) (*GroupDetail, error) {
	raw, _, err := c.do(http.MethodPost, "/aa/group/show", map[string]int64{"id": id})
	if err != nil {
		return nil, err
	}
	var g GroupDetail
	if err := json.Unmarshal(raw, &g); err != nil {
		return nil, fmt.Errorf("decode group: %w (body=%s)", err, truncate(string(raw), 200))
	}
	return &g, nil
}

// ResolveGroupID returns the group id matching target (numeric id or name).
func (c *MgtClient) ResolveGroupID(target string) (int64, error) {
	if id, err := strconv.ParseInt(target, 10, 64); err == nil {
		return id, nil
	}
	gs, err := c.ListGroups()
	if err != nil {
		return 0, err
	}
	for _, g := range gs {
		if g.Name == target {
			return g.ID, nil
		}
	}
	return 0, fmt.Errorf("no group with name or id %q", target)
}

func (c *MgtClient) AssignUserToGroup(username string, groupID int64) error {
	_, _, err := c.do(http.MethodPost, "/aa/group/user/assign", map[string]any{
		"username": username,
		"group_id": groupID,
	})
	return err
}

func (c *MgtClient) UnassignUserFromGroup(username string, groupID int64) error {
	_, _, err := c.do(http.MethodPost, "/aa/group/user/unassign", map[string]any{
		"username": username,
		"group_id": groupID,
	})
	return err
}

func (c *MgtClient) AssignNEToGroup(groupID, neID int64) error {
	_, _, err := c.do(http.MethodPost, "/aa/group/ne/assign", map[string]int64{
		"group_id": groupID,
		"ne_id":    neID,
	})
	return err
}

func (c *MgtClient) UnassignNEFromGroup(groupID, neID int64) error {
	_, _, err := c.do(http.MethodPost, "/aa/group/ne/unassign", map[string]int64{
		"group_id": groupID,
		"ne_id":    neID,
	})
	return err
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
