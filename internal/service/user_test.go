package service_test

import (
	"errors"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/internal/service"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/store"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/testutil"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// ── AddUser ───────────────────────────────────────────────────────────────────

func TestAddUser_Success(t *testing.T) {
	var got *db_models.TblAccount
	store.SetSingleton(&testutil.MockStore{
		AddUserFn: func(account *db_models.TblAccount) error {
			got = account
			return nil
		},
	})

	u := &db_models.TblAccount{AccountID: 1, AccountName: "alice"}
	if err := service.AddUser(u); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != u {
		t.Error("AddUser did not pass the correct account to store")
	}
}

func TestAddUser_DBError(t *testing.T) {
	dbErr := errors.New("insert failed")
	store.SetSingleton(&testutil.MockStore{
		AddUserFn: func(_ *db_models.TblAccount) error { return dbErr },
	})

	if err := service.AddUser(&db_models.TblAccount{}); !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── UpdateUser ────────────────────────────────────────────────────────────────

func TestUpdateUser_Success(t *testing.T) {
	var got *db_models.TblAccount
	store.SetSingleton(&testutil.MockStore{
		UpdateUserFn: func(account *db_models.TblAccount) error {
			got = account
			return nil
		},
	})

	u := &db_models.TblAccount{AccountID: 2, AccountName: "bob"}
	if err := service.UpdateUser(u); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != u {
		t.Error("UpdateUser did not pass the correct account to store")
	}
}

func TestUpdateUser_DBError(t *testing.T) {
	dbErr := errors.New("update failed")
	store.SetSingleton(&testutil.MockStore{
		UpdateUserFn: func(_ *db_models.TblAccount) error { return dbErr },
	})

	if err := service.UpdateUser(&db_models.TblAccount{}); !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── GetUserByUserName ─────────────────────────────────────────────────────────

func TestGetUserByUserName_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 3, AccountName: name}, nil
		},
	})

	u, err := service.GetUserByUserName("alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.AccountName != "alice" {
		t.Errorf("AccountName: got %q, want %q", u.AccountName, "alice")
	}
	if u.AccountID != 3 {
		t.Errorf("AccountID: got %d, want 3", u.AccountID)
	}
}

func TestGetUserByUserName_NotFound(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(_ string) (*db_models.TblAccount, error) {
			return nil, nil
		},
	})

	u, err := service.GetUserByUserName("ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u != nil {
		t.Errorf("expected nil, got %v", u)
	}
}

func TestGetUserByUserName_DBError(t *testing.T) {
	dbErr := errors.New("query failed")
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(_ string) (*db_models.TblAccount, error) {
			return nil, dbErr
		},
	})

	_, err := service.GetUserByUserName("alice")
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── GetAllUser ────────────────────────────────────────────────────────────────

func TestGetAllUser_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{
				{AccountID: 1, AccountName: "alice"},
				{AccountID: 2, AccountName: "bob"},
			}, nil
		},
	})

	users, err := service.GetAllUser()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("user count: got %d, want 2", len(users))
	}
}

func TestGetAllUser_Empty(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return []*db_models.TblAccount{}, nil
		},
	})

	users, err := service.GetAllUser()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected empty list, got %d users", len(users))
	}
}

func TestGetAllUser_DBError(t *testing.T) {
	dbErr := errors.New("connection refused")
	store.SetSingleton(&testutil.MockStore{
		GetAllUserFn: func() ([]*db_models.TblAccount, error) {
			return nil, dbErr
		},
	})

	_, err := service.GetAllUser()
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}
