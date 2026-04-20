package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// ── HandlerAdminUserUpdate ────────────────────────────────────────────────────

func TestHandlerAdminUserUpdate_Success(t *testing.T) {
	var saved *db_models.TblAccount
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 5, AccountName: name, AccountType: 2, Email: "old@x.com"}, nil
		},
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{
				{AccountID: 5, AccountName: "bob", Email: "old@x.com"},
			}, nil
		},
		GetGroupByIdFn: func(id int64) (*db_models.CliGroup, error) {
			return &db_models.CliGroup{ID: id, Name: "ops"}, nil
		},
		GetAllGroupsOfUserFn: func(userId int64) ([]*db_models.CliUserGroupMapping, error) {
			return []*db_models.CliUserGroupMapping{{UserID: userId, GroupID: 1}}, nil
		},
		UpdateUserFn: func(u *db_models.TblAccount) error {
			saved = u
			return nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]any{
		"account_name": "bob",
		"full_name":    "Bob Doe",
		"email":        "bob@example.com",
		"phone_number": "0900000000",
		"address":      "HCM",
		"description":  "ops",
		"account_type": 1,
		"group_ids":    []int64{1},
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminUserUpdate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	if saved == nil {
		t.Fatal("UpdateUser was not called")
	}
	if saved.FullName != "Bob Doe" || saved.Email != "bob@example.com" || saved.PhoneNumber != "0900000000" ||
		saved.Address != "HCM" || saved.Description != "ops" || saved.AccountType != 1 {
		t.Errorf("updated fields wrong: %+v", saved)
	}
}

func TestHandlerAdminUserUpdate_MissingAccountName(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	body, _ := json.Marshal(map[string]string{"full_name": "x"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminUserUpdate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerAdminUserUpdate_UserNotFound(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) { return nil, nil },
	})

	body, _ := json.Marshal(map[string]any{"account_name": "ghost", "account_type": 2})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminUserUpdate(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", w.Code)
	}
}

func TestHandlerAdminUserUpdate_RejectsSuperAdminTarget(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 1, AccountName: name, AccountType: 0}, nil
		},
	})

	body, _ := json.Marshal(map[string]any{"account_name": "root", "account_type": 1})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminUserUpdate(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403 (cannot edit SuperAdmin)", w.Code)
	}
}

func TestHandlerAdminUserUpdate_RejectsElevateToSuperAdmin(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 5, AccountName: name, AccountType: 2}, nil
		},
	})

	body, _ := json.Marshal(map[string]any{"account_name": "bob", "account_type": 0})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminUserUpdate(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403 (cannot elevate to SuperAdmin)", w.Code)
	}
}

func TestHandlerAdminUserUpdate_DBError(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 5, AccountName: name, AccountType: 2}, nil
		},
		UpdateUserFn:         func(u *db_models.TblAccount) error { return errors.New("db down") },
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]any{"account_name": "bob", "account_type": 2})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminUserUpdate(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

// ── HandlerAdminUserList ──────────────────────────────────────────────────────

func TestHandlerAdminUserList_ExcludesPassword(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{
				{AccountID: 1, AccountName: "alice", Password: "SECRET-HASH", Email: "a@x.com", AccountType: 1, IsEnable: true},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminUserList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte("alice")) {
		t.Errorf("response missing username: %s", body)
	}
	if bytes.Contains([]byte(body), []byte("SECRET-HASH")) {
		t.Errorf("response leaked password hash: %s", body)
	}
	if bytes.Contains([]byte(body), []byte("\"password\"")) {
		t.Errorf("response should not include password field: %s", body)
	}
}

func TestHandlerAdminUserList_EmptyReturnsEmptyArray(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllUserFn: func() ([]*db_models.TblAccount, error) { return nil, nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminUserList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	body := bytes.TrimSpace(w.Body.Bytes())
	if string(body) != "[]" {
		t.Errorf("expected empty JSON array, got %q", string(body))
	}
}
