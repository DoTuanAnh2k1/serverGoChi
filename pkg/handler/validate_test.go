package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/token"
)

// ── HandlerValidateToken ──────────────────────────────────────────────────────

func makeValidateBody(t *testing.T, tok string) *bytes.Buffer {
	t.Helper()
	b, _ := json.Marshal(map[string]string{"token": tok})
	return bytes.NewBuffer(b)
}

func TestHandlerValidateToken_ValidToken(t *testing.T) {
	tok, err := token.CreateToken("alice", "admin viewer")
	if err != nil {
		t.Fatalf("CreateToken: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/aa/validate-token", makeValidateBody(t, tok))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandlerValidateToken(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	var resp struct {
		Username string `json:"username"`
		Roles    string `json:"roles"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Username != "alice" {
		t.Errorf("username: got %q, want %q", resp.Username, "alice")
	}
	if resp.Roles != "admin viewer" {
		t.Errorf("roles: got %q, want %q", resp.Roles, "admin viewer")
	}
}

func TestHandlerValidateToken_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/aa/validate-token",
		makeValidateBody(t, "Basic not.a.valid.token"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandlerValidateToken(w, req)

	if w.Code == http.StatusOK {
		t.Error("invalid token should not return 200")
	}
}

func TestHandlerValidateToken_MissingBasicPrefix(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/aa/validate-token",
		makeValidateBody(t, "just-a-raw-string"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandlerValidateToken(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", w.Code)
	}
}

func TestHandlerValidateToken_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/aa/validate-token",
		strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.HandlerValidateToken(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

func TestHandlerValidateToken_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/aa/validate-token", nil)
	w := httptest.NewRecorder()

	handler.HandlerValidateToken(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want 405", w.Code)
	}
}
