package service_test

import (
	"strings"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

// A disabled account's email should NOT block a fresh account from claiming
// that email — the rule is "unique among active users".
func TestEnsureEmailUnique_SkipsDisabledAccounts(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{
				{AccountName: "olduser", Email: "shared@example.com", IsEnable: false},
			}, nil
		},
	})

	// New user wants "shared@example.com" — olduser is disabled so allowed.
	if err := service.EnsureEmailUnique("shared@example.com", "newuser"); err != nil {
		t.Errorf("expected nil (disabled user's email is free), got %v", err)
	}
}

// An active account's email still blocks other accounts.
func TestEnsureEmailUnique_BlocksOnActive(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{
				{AccountName: "alice", Email: "shared@example.com", IsEnable: true},
			}, nil
		},
	})

	err := service.EnsureEmailUnique("shared@example.com", "newuser")
	if err == nil || !strings.Contains(err.Error(), "already in use") {
		t.Errorf("expected 'already in use', got %v", err)
	}
}

// excludeAccountName skip still works so a user can keep their own email.
func TestEnsureEmailUnique_SkipsSelf(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{
				{AccountName: "alice", Email: "a@example.com", IsEnable: true},
			}, nil
		},
	})

	if err := service.EnsureEmailUnique("a@example.com", "alice"); err != nil {
		t.Errorf("self-owned email must pass, got %v", err)
	}
}
