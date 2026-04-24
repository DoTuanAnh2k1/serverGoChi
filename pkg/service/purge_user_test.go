package service_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

// PurgeUser must hit every cascade in order: user-group mappings, user-ne
// mappings, password history, then finally the account row.
func TestPurgeUser_FullCascade(t *testing.T) {
	var order []string
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 42, AccountName: name, AccountType: 2}, nil
		},
		DeleteAllUserGroupMappingByUserIdFn: func(uid int64) error {
			order = append(order, "group"); return nil
		},
		DeleteAllUserNeMappingByUserIdFn: func(uid int64) error {
			order = append(order, "ne"); return nil
		},
		PrunePasswordHistoryFn: func(uid int64, keep int) error {
			if keep != 0 { t.Errorf("expected keep=0 to delete all history, got %d", keep) }
			order = append(order, "history"); return nil
		},
		DeleteUserByIdFn: func(id int64) error {
			if id != 42 { t.Errorf("DeleteUserById id=%d want 42", id) }
			order = append(order, "account"); return nil
		},
	})

	if err := service.PurgeUser("alice"); err != nil {
		t.Fatalf("err: %v", err)
	}
	want := []string{"group", "ne", "history", "account"}
	if len(order) != len(want) {
		t.Fatalf("cascade calls: got %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("step %d: got %q want %q (full seq: %v)", i, order[i], want[i], order)
		}
	}
}

func TestPurgeUser_RefusesSuperAdmin(t *testing.T) {
	accountDeleted := false
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 1, AccountName: name, AccountType: 0}, nil
		},
		DeleteUserByIdFn: func(int64) error { accountDeleted = true; return nil },
	})
	err := service.PurgeUser("superadmin")
	if err == nil || !strings.Contains(err.Error(), "SuperAdmin") {
		t.Errorf("expected SuperAdmin refusal, got %v", err)
	}
	if accountDeleted {
		t.Errorf("must not touch tbl_account when target is SuperAdmin")
	}
}

func TestPurgeUser_NotFound(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(string) (*db_models.TblAccount, error) { return nil, nil },
	})
	err := service.PurgeUser("ghost")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not-found error, got %v", err)
	}
}

// Cascade failures are logged but don't abort the purge — the final
// DeleteUserById is authoritative. This keeps the account from getting
// stuck disabled-but-undeletable when a mapping table is flaky.
func TestPurgeUser_CascadeFailuresNonFatal(t *testing.T) {
	accountDeleted := false
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 9, AccountName: "bob", AccountType: 2}, nil
		},
		DeleteAllUserGroupMappingByUserIdFn: func(int64) error { return errors.New("flaky") },
		DeleteAllUserNeMappingByUserIdFn:    func(int64) error { return errors.New("flaky") },
		PrunePasswordHistoryFn:              func(int64, int) error { return errors.New("flaky") },
		DeleteUserByIdFn:                    func(int64) error { accountDeleted = true; return nil },
	})
	if err := service.PurgeUser("bob"); err != nil {
		t.Fatalf("cascade flaky must not block purge, got %v", err)
	}
	if !accountDeleted {
		t.Errorf("tbl_account row must be deleted even if cascades fail")
	}
}

// When the final DeleteUserById itself fails, the error propagates.
func TestPurgeUser_AccountDeleteErrorPropagates(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 9, AccountName: "bob", AccountType: 2}, nil
		},
		DeleteUserByIdFn: func(int64) error { return errors.New("boom") },
	})
	err := service.PurgeUser("bob")
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("DeleteUserById error should propagate, got %v", err)
	}
}
