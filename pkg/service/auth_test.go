package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
	appcrypt "github.com/DoTuanAnh2k1/serverGoChi/pkg/bcrypt"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// ── Authenticate ──────────────────────────────────────────────────────────────

func TestAuthenticate_Success(t *testing.T) {
	hash := appcrypt.Encode("alice" + "secret")
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 7, IsEnable: true, Password: hash}, nil
		},
	})

	ok, err, id := service.Authenticate("alice", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("expected true for valid credentials")
	}
	if id != 7 {
		t.Errorf("accountID: got %d, want 7", id)
	}
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	hash := appcrypt.Encode("alice" + "correct")
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 1, IsEnable: true, Password: hash}, nil
		},
	})

	ok, err, _ := service.Authenticate("alice", "wrong")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false for wrong password")
	}
}

func TestAuthenticate_UserDisabled(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 1, IsEnable: false, Password: "hash"}, nil
		},
	})

	ok, err, id := service.Authenticate("alice", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected false for disabled user")
	}
	if id != -1 {
		t.Errorf("expected id=-1, got %d", id)
	}
}

func TestAuthenticate_DBError(t *testing.T) {
	dbErr := errors.New("connection refused")
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(name string) (*db_models.TblAccount, error) {
			return nil, dbErr
		},
	})

	ok, err, _ := service.Authenticate("alice", "password")
	if ok {
		t.Error("expected false on DB error")
	}
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── GetRolesById ─────────────────────────────────────────────────────────────

func TestGetRolesById_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetRolesByIdFn: func(userID int64) ([]*db_models.CliRoleUserMapping, error) {
			return []*db_models.CliRoleUserMapping{
				{Permission: "admin"},
				{Permission: "viewer"},
			}, nil
		},
	})

	roles, err := service.GetRolesById(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, want := range []string{"admin", "viewer"} {
		if !containsWord(roles, want) {
			t.Errorf("role %q not found in %q", want, roles)
		}
	}
}

func TestGetRolesById_DBError(t *testing.T) {
	dbErr := errors.New("query failed")
	store.SetSingleton(&testutil.MockStore{
		GetRolesByIdFn: func(userID int64) ([]*db_models.CliRoleUserMapping, error) {
			return nil, dbErr
		},
	})

	_, err := service.GetRolesById(1)
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

func TestGetRolesById_NoRoles(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetRolesByIdFn: func(userID int64) ([]*db_models.CliRoleUserMapping, error) {
			return []*db_models.CliRoleUserMapping{}, nil
		},
	})

	roles, err := service.GetRolesById(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = roles // empty is valid
}

// ── UpdateLoginHistory ────────────────────────────────────────────────────────

func TestUpdateLoginHistory_Success(t *testing.T) {
	var capturedUser, capturedIP string
	var capturedTime time.Time
	store.SetSingleton(&testutil.MockStore{
		UpdateLoginHistoryFn: func(username, ip string, t time.Time) error {
			capturedUser = username
			capturedIP = ip
			capturedTime = t
			return nil
		},
	})

	if err := service.UpdateLoginHistory("alice", "10.0.0.1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedUser != "alice" {
		t.Errorf("username: got %q, want %q", capturedUser, "alice")
	}
	if capturedIP != "10.0.0.1" {
		t.Errorf("ip: got %q, want %q", capturedIP, "10.0.0.1")
	}
	if capturedTime.IsZero() {
		t.Error("time should not be zero")
	}
}

func TestUpdateLoginHistory_DBError(t *testing.T) {
	dbErr := errors.New("insert failed")
	store.SetSingleton(&testutil.MockStore{
		UpdateLoginHistoryFn: func(username, ip string, t time.Time) error {
			return dbErr
		},
	})

	if err := service.UpdateLoginHistory("bob", "127.0.0.1"); !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── GetTblIdByUserId ──────────────────────────────────────────────────────────

func TestGetTblIdByUserId_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCLIUserNeMappingByUserIdFn: func(userID int64) (*db_models.CliUserNeMapping, error) {
			return &db_models.CliUserNeMapping{UserID: userID, TblNeID: 42}, nil
		},
	})

	id, err := service.GetTblIdByUserId(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Errorf("TblNeID: got %d, want 42", id)
	}
}

// ── GetNeListById ─────────────────────────────────────────────────────────────

func TestGetNeListById_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetNeListByIdFn: func(id int64) ([]*db_models.CliNe, error) {
			return []*db_models.CliNe{{ID: id, Name: "NE-HCM-01"}}, nil
		},
	})

	list, err := service.GetNeListById(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 1 || list[0].Name != "NE-HCM-01" {
		t.Errorf("unexpected result: %v", list)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func containsWord(s, word string) bool {
	for i := 0; i+len(word) <= len(s); i++ {
		if s[i:i+len(word)] == word {
			before := i == 0 || s[i-1] == ' '
			after := i+len(word) == len(s) || s[i+len(word)] == ' '
			if before && after {
				return true
			}
		}
	}
	return false
}
