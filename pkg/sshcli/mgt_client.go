package sshcli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// mgtClient is a minimal HTTP client for cli-mgt-svc v2.
type mgtClient struct {
	base    string
	token   string
	role    string
	http    *http.Client
}

func newMgtClient(base string, timeout time.Duration) *mgtClient {
	return &mgtClient{
		base: base,
		http: &http.Client{Timeout: timeout},
	}
}

func (c *mgtClient) Authenticate(username, password string) error {
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	var resp struct {
		Status string `json:"status"`
		Token  string `json:"token"`
		Role   string `json:"role"`
	}
	if err := c.post("/aa/authenticate", body, false, &resp); err != nil {
		return err
	}
	c.token = resp.Token
	c.role = resp.Role
	return nil
}

// NE is the subset of fields needed by the gate.
type NE struct {
	ID          int64  `json:"id"`
	Namespace   string `json:"namespace"`
	NeType      string `json:"ne_type"`
	SiteName    string `json:"site_name"`
	MasterIP    string `json:"master_ip"`
	MasterPort  int32  `json:"master_port"`
	SSHUsername string `json:"ssh_username"`
	SSHPassword string `json:"ssh_password"`
	ConfMode    string `json:"conf_mode"`
}

func (c *mgtClient) ListNEs() ([]NE, error) {
	var out []NE
	err := c.get("/aa/nes", &out)
	return out, err
}

func (c *mgtClient) GetNEByNamespace(ns string) (*NE, error) {
	nes, err := c.ListNEs()
	if err != nil {
		return nil, err
	}
	for i := range nes {
		if nes[i].Namespace == ns {
			return &nes[i], nil
		}
	}
	return nil, nil
}

// AuthorizeCheck calls POST /aa/authorize/check and returns (allowed, reason, error).
func (c *mgtClient) AuthorizeCheck(username string, neID, commandID int64) (bool, string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"username":   username,
		"ne_id":      neID,
		"command_id": commandID,
	})
	var resp struct {
		Allowed bool   `json:"allowed"`
		Reason  string `json:"reason"`
	}
	if err := c.post("/aa/authorize/check", body, true, &resp); err != nil {
		return false, "", err
	}
	return resp.Allowed, resp.Reason, nil
}

func (c *mgtClient) SaveHistory(account, cmdText, neNamespace, neIP, scope, result string) error {
	body, _ := json.Marshal(map[string]string{
		"account":      account,
		"cmd_text":     cmdText,
		"ne_namespace": neNamespace,
		"ne_ip":        neIP,
		"scope":        scope,
		"result":       result,
	})
	// /aa/history/save is unauthenticated — skip token.
	return c.postRaw("/aa/history/save", body, false)
}

func (c *mgtClient) get(path string, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, c.base+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("GET %s: %w", path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GET %s: HTTP %d: %s", path, resp.StatusCode, b)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *mgtClient) post(path string, body []byte, withToken bool, out interface{}) error {
	req, err := http.NewRequest(http.MethodPost, c.base+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if withToken {
		req.Header.Set("Authorization", c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("POST %s: %w", path, err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("POST %s: HTTP %d: %s", path, resp.StatusCode, b)
	}
	if out != nil {
		return json.Unmarshal(b, out)
	}
	return nil
}

func (c *mgtClient) postRaw(path string, body []byte, withToken bool) error {
	return c.post(path, body, withToken, nil)
}
