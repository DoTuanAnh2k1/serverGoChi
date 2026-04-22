package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

func postUserSet(t *testing.T, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/aa/authenticate/user/set", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()
	handler.HandlerAuthenticateUserSet(w, req)
	return w
}

// Re-enable a disabled user: non-empty fields from the request overwrite the
// stored values; fields omitted from the request keep their existing values.
func TestHandlerUserSet_ReEnable_MergesNonEmptyFields(t *testing.T) {
	stored := &db_models.TblAccount{
		AccountID:   7,
		AccountName: "alice",
		Email:       "old@example.com",
		FullName:    "Alice Old",
		PhoneNumber: "0900000000",
		Address:     "old address",
		Description: "was disabled",
		AccountType: 2,
		IsEnable:    false,
		Password:    "oldhash",
	}
	var updated *db_models.TblAccount
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			if name == "alice" {
				return stored, nil
			}
			return nil, nil
		},
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{stored}, nil
		},
		UpdateUserFn: func(a *db_models.TblAccount) error {
			updated = a
			return nil
		},
		SaveHistoryCommandFn: func(_ db_models.CliOperationHistory) error { return nil },
	})

	// Request supplies new email + full_name, omits phone/address/description.
	w := postUserSet(t, map[string]any{
		"account_name": "alice",
		"password":     "newpass",
		"email":        "new@example.com",
		"full_name":    "Alice New",
		"account_type": 2,
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want 201 (body=%s)", w.Code, w.Body.String())
	}
	if updated == nil {
		t.Fatal("UpdateUser not called")
	}
	if !updated.IsEnable {
		t.Errorf("IsEnable: got %v, want true", updated.IsEnable)
	}
	if updated.Email != "new@example.com" {
		t.Errorf("Email: got %q, want new", updated.Email)
	}
	if updated.FullName != "Alice New" {
		t.Errorf("FullName: got %q, want Alice New", updated.FullName)
	}
	// Omitted fields keep their stored values.
	if updated.PhoneNumber != "0900000000" {
		t.Errorf("PhoneNumber: expected preserved, got %q", updated.PhoneNumber)
	}
	if updated.Address != "old address" {
		t.Errorf("Address: expected preserved, got %q", updated.Address)
	}
	if updated.Description != "was disabled" {
		t.Errorf("Description: expected preserved, got %q", updated.Description)
	}
	if updated.Password == "oldhash" {
		t.Errorf("Password should be refreshed on re-enable, still %q", updated.Password)
	}
}

// Already-active users stay 304 — re-enable merge does not apply.
func TestHandlerUserSet_EnabledUser_ReturnsNotModified(t *testing.T) {
	stored := &db_models.TblAccount{
		AccountID:   1,
		AccountName: "alice",
		AccountType: 2,
		IsEnable:    true,
	}
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn:  func(_ string) (*db_models.TblAccount, error) { return stored, nil },
		GetAllUserFn:         func() ([]*db_models.TblAccount, error) { return []*db_models.TblAccount{stored}, nil },
		UpdateUserFn:         func(_ *db_models.TblAccount) error { t.Error("UpdateUser should not be called"); return nil },
		SaveHistoryCommandFn: func(_ db_models.CliOperationHistory) error { return nil },
	})

	w := postUserSet(t, map[string]any{
		"account_name": "alice",
		"password":     "anything",
		"account_type": 2,
	})
	if w.Code != http.StatusNotModified {
		t.Errorf("status: got %d, want 304", w.Code)
	}
}

// Fresh username but email collides with a DISABLED account — should succeed,
// because disabled accounts' emails are considered free.
func TestHandlerUserSet_NewUser_CanReuseDisabledEmail(t *testing.T) {
	disabled := &db_models.TblAccount{
		AccountID:   1,
		AccountName: "olduser",
		Email:       "shared@example.com",
		IsEnable:    false,
		AccountType: 2,
	}
	var added *db_models.TblAccount
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			if name == "olduser" {
				return disabled, nil
			}
			return nil, nil
		},
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{disabled}, nil
		},
		AddUserFn: func(a *db_models.TblAccount) error {
			added = a
			return nil
		},
		SaveHistoryCommandFn: func(_ db_models.CliOperationHistory) error { return nil },
	})

	w := postUserSet(t, map[string]any{
		"account_name": "newuser",
		"password":     "pw",
		"email":        "shared@example.com",
		"account_type": 2,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want 201 (body=%s)", w.Code, w.Body.String())
	}
	if added == nil || added.Email != "shared@example.com" {
		t.Errorf("expected new user with reused email, got %+v", added)
	}
}
