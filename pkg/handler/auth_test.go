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

// TestHandlerAuthenticateUserDelete_RejectsSeedUser ensures the system user
// cannot be disabled via the API regardless of the caller's permission level.
func TestHandlerAuthenticateUserDelete_RejectsSeedUser(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]string{"account_name": "anhdt195"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAuthenticateUserDelete(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("status: got %d, want 403 (seed user cannot be disabled)", w.Code)
	}
}

// TestHandlerAuthenticateUserDelete_DisablesNormalUser ensures regular users
// can still be disabled normally.
func TestHandlerAuthenticateUserDelete_DisablesNormalUser(t *testing.T) {
	var disabled bool
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 7, AccountName: name, IsEnable: true}, nil
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
