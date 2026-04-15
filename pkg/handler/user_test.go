package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// injectUser creates a request with an authenticated admin user in context.
func injectUser(r *http.Request, username string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserContextKey, &middleware.User{
		Username:   username,
		Permission: "admin",
	})
	return r.WithContext(ctx)
}

// ── HandlerAuthorizeUserSet ───────────────────────────────────────────────────

func TestHandlerAuthorizeUserSet_SetAdmin(t *testing.T) {
	var updatedUser *db_models.TblAccount
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 5, AccountName: name, AccountType: 2}, nil
		},
		UpdateUserFn: func(u *db_models.TblAccount) error {
			updatedUser = u
			return nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]string{"username": "bob", "permission": "admin"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAuthorizeUserSet(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	if updatedUser == nil || updatedUser.AccountType != 1 {
		t.Errorf("account_type: got %v, want 1", updatedUser)
	}
}

func TestHandlerAuthorizeUserSet_SetUser(t *testing.T) {
	var updatedUser *db_models.TblAccount
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 5, AccountName: name, AccountType: 1}, nil
		},
		UpdateUserFn: func(u *db_models.TblAccount) error {
			updatedUser = u
			return nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]string{"username": "bob", "permission": "user"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAuthorizeUserSet(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	if updatedUser == nil || updatedUser.AccountType != 2 {
		t.Errorf("account_type: got %v, want 2", updatedUser)
	}
}

func TestHandlerAuthorizeUserSet_InvalidPermission(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 5, AccountName: name}, nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]string{"username": "bob", "permission": "superadmin"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAuthorizeUserSet(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerAuthorizeUserSet_UserNotFound(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return nil, nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]string{"username": "nobody", "permission": "admin"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAuthorizeUserSet(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", w.Code)
	}
}

// ── HandlerAuthorizeUserDelete ────────────────────────────────────────────────

func TestHandlerAuthorizeUserDelete_ResetsToUser(t *testing.T) {
	var updatedUser *db_models.TblAccount
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 5, AccountName: name, AccountType: 1}, nil
		},
		UpdateUserFn: func(u *db_models.TblAccount) error {
			updatedUser = u
			return nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]string{"username": "bob"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAuthorizeUserDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	if updatedUser == nil || updatedUser.AccountType != 2 {
		t.Errorf("account_type: got %v, want 2 (reset to user)", updatedUser)
	}
}

func TestHandlerAuthorizeUserDelete_UserNotFound(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return nil, nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]string{"username": "nobody"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAuthorizeUserDelete(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", w.Code)
	}
}

func TestHandlerAuthorizeUserSet_RejectsSuperAdmin(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 1, AccountName: name, AccountType: 0}, nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]string{"username": "superadmin", "permission": "user"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAuthorizeUserSet(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403 (SuperAdmin cannot have permission changed)", w.Code)
	}
}

func TestHandlerAuthorizeUserDelete_RejectsSuperAdmin(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 1, AccountName: name, AccountType: 0}, nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]string{"username": "superadmin"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAuthorizeUserDelete(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403 (SuperAdmin permission cannot be reset)", w.Code)
	}
}

// ── HandlerAuthorizeUserShow ──────────────────────────────────────────────────

func TestHandlerAuthorizeUserShow_ReturnsPermissions(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{
				{AccountID: 1, AccountName: "alice", AccountType: 0},
				{AccountID: 2, AccountName: "bob", AccountType: 2},
			}, nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAuthorizeUserShow(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("status: got %d, want 302", w.Code)
	}

	var result []map[string]string
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}

	perms := map[string]string{}
	for _, r := range result {
		perms[r["username"]] = r["permission"]
	}
	if perms["alice"] != "admin" {
		t.Errorf("alice permission: got %q, want %q", perms["alice"], "admin")
	}
	if perms["bob"] != "user" {
		t.Errorf("bob permission: got %q, want %q", perms["bob"], "user")
	}
}
