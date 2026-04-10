package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/internal/handler"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/store"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/testutil"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

func reqWithUser(r *http.Request, u *middleware.User) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserContextKey, u)
	return r.WithContext(ctx)
}

func makeHistoryBody(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

// ── HandlerSaveHistory ────────────────────────────────────────────────────────

func TestHandlerSaveHistory_Success(t *testing.T) {
	var saved db_models.CliOperationHistory
	store.SetSingleton(&testutil.MockStore{
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error {
			saved = h
			return nil
		},
	})

	body := makeHistoryBody(t, map[string]interface{}{
		"cmd_name": "show version",
		"ne_name":  "NE-HCM-01",
		"ne_ip":    "10.0.0.1",
		"result":   "success",
	})
	req := httptest.NewRequest(http.MethodPost, "/aa/history/save", body)
	req.Header.Set("Content-Type", "application/json")
	req = reqWithUser(req, &middleware.User{Username: "alice", Roles: "admin"})
	w := httptest.NewRecorder()

	handler.HandlerSaveHistory(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status: got %d, want 201", w.Code)
	}
	if saved.CmdName != "show version" {
		t.Errorf("CmdName: got %q, want %q", saved.CmdName, "show version")
	}
	if saved.Account != "alice" {
		t.Errorf("Account: got %q, want %q", saved.Account, "alice")
	}
}

func TestHandlerSaveHistory_MissingCmdName(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	body := makeHistoryBody(t, map[string]interface{}{
		"cmd_name": "",
		"ne_name":  "NE-HCM-01",
	})
	req := httptest.NewRequest(http.MethodPost, "/aa/history/save", body)
	req.Header.Set("Content-Type", "application/json")
	req = reqWithUser(req, &middleware.User{Username: "alice", Roles: "admin"})
	w := httptest.NewRecorder()

	handler.HandlerSaveHistory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerSaveHistory_MissingNeName(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	body := makeHistoryBody(t, map[string]interface{}{
		"cmd_name": "show ip route",
		"ne_name":  "   ",
	})
	req := httptest.NewRequest(http.MethodPost, "/aa/history/save", body)
	req.Header.Set("Content-Type", "application/json")
	req = reqWithUser(req, &middleware.User{Username: "bob", Roles: "admin"})
	w := httptest.NewRecorder()

	handler.HandlerSaveHistory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerSaveHistory_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/aa/history/save",
		strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	req = reqWithUser(req, &middleware.User{Username: "alice", Roles: "admin"})
	w := httptest.NewRecorder()

	handler.HandlerSaveHistory(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerSaveHistory_DBError(t *testing.T) {
	dbErr := errors.New("insert failed")
	store.SetSingleton(&testutil.MockStore{
		SaveHistoryCommandFn: func(_ db_models.CliOperationHistory) error { return dbErr },
	})

	body := makeHistoryBody(t, map[string]interface{}{
		"cmd_name": "show version",
		"ne_name":  "NE-01",
	})
	req := httptest.NewRequest(http.MethodPost, "/aa/history/save", body)
	req.Header.Set("Content-Type", "application/json")
	req = reqWithUser(req, &middleware.User{Username: "alice", Roles: "admin"})
	w := httptest.NewRecorder()

	handler.HandlerSaveHistory(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

func TestHandlerSaveHistory_NoUserInContext(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	body := makeHistoryBody(t, map[string]interface{}{
		"cmd_name": "show version",
		"ne_name":  "NE-01",
	})
	req := httptest.NewRequest(http.MethodPost, "/aa/history/save", body)
	req.Header.Set("Content-Type", "application/json")
	// no user injected
	w := httptest.NewRecorder()

	handler.HandlerSaveHistory(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}
