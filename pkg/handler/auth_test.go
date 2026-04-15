package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// TestHandlerAuthenticateUserDelete_RejectsSuperAdmin ensures accounts with
// account_type=0 cannot be disabled via the API, regardless of caller.
func TestHandlerAuthenticateUserDelete_RejectsSuperAdmin(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 1, AccountName: name, AccountType: 0, IsEnable: true}, nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]string{"account_name": "superadmin"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAuthenticateUserDelete(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403 (SuperAdmin cannot be disabled)", w.Code)
	}
}

// TestHandlerAuthenticateUserDelete_DisablesNormalUser ensures regular users
// (account_type >= 1) can still be disabled normally.
func TestHandlerAuthenticateUserDelete_DisablesNormalUser(t *testing.T) {
	var disabled bool
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 7, AccountName: name, AccountType: 2, IsEnable: true}, nil
		},
		UpdateUserFn: func(u *db_models.TblAccount) error {
			disabled = !u.IsEnable
			return nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]string{"account_name": "bob"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAuthenticateUserDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	if !disabled {
		t.Error("expected user to be disabled")
	}
}
