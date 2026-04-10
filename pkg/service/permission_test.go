package service_test

import (
	"errors"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// ── IsExistCliRole ────────────────────────────────────────────────────────────

func TestIsExistCliRole_Exists(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliRoleFn: func(role *db_models.CliRole) (*db_models.CliRole, error) {
			return &db_models.CliRole{RoleID: 1, Permission: role.Permission}, nil
		},
	})

	exists, err := service.IsExistCliRole(&db_models.CliRole{Permission: "admin"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected true, role should exist")
	}
}

func TestIsExistCliRole_NotExists(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliRoleFn: func(_ *db_models.CliRole) (*db_models.CliRole, error) {
			return nil, nil
		},
	})

	exists, err := service.IsExistCliRole(&db_models.CliRole{Permission: "ghost"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false, role should not exist")
	}
}

func TestIsExistCliRole_DBError(t *testing.T) {
	dbErr := errors.New("query failed")
	store.SetSingleton(&testutil.MockStore{
		GetCliRoleFn: func(_ *db_models.CliRole) (*db_models.CliRole, error) {
			return nil, dbErr
		},
	})

	_, err := service.IsExistCliRole(&db_models.CliRole{Permission: "admin"})
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── CreateCliRole ─────────────────────────────────────────────────────────────

func TestCreateCliRole_Success(t *testing.T) {
	var got *db_models.CliRole
	store.SetSingleton(&testutil.MockStore{
		CreateCliRoleFn: func(role *db_models.CliRole) error {
			got = role
			return nil
		},
	})

	r := &db_models.CliRole{Permission: "operator", Scope: "read"}
	if err := service.CreateCliRole(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != r {
		t.Error("CreateCliRole did not pass the correct role to store")
	}
}

func TestCreateCliRole_DBError(t *testing.T) {
	dbErr := errors.New("insert failed")
	store.SetSingleton(&testutil.MockStore{
		CreateCliRoleFn: func(_ *db_models.CliRole) error { return dbErr },
	})

	if err := service.CreateCliRole(&db_models.CliRole{}); !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── DeleteCliRole ─────────────────────────────────────────────────────────────

func TestDeleteCliRole_Success(t *testing.T) {
	var deleted *db_models.CliRole
	store.SetSingleton(&testutil.MockStore{
		DeleteCliRoleFn: func(role *db_models.CliRole) error {
			deleted = role
			return nil
		},
	})

	r := &db_models.CliRole{RoleID: 5, Permission: "viewer"}
	if err := service.DeleteCliRole(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != r {
		t.Error("DeleteCliRole did not pass the correct role to store")
	}
}

func TestDeleteCliRole_DBError(t *testing.T) {
	dbErr := errors.New("delete failed")
	store.SetSingleton(&testutil.MockStore{
		DeleteCliRoleFn: func(_ *db_models.CliRole) error { return dbErr },
	})

	if err := service.DeleteCliRole(&db_models.CliRole{}); !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── GetAllCliRoles ────────────────────────────────────────────────────────────

func TestGetAllCliRoles_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllCliRoleFn: func() ([]*db_models.CliRole, error) {
			return []*db_models.CliRole{
				{RoleID: 1, Permission: "admin"},
				{RoleID: 2, Permission: "viewer"},
			}, nil
		},
	})

	roles, err := service.GetAllCliRoles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("role count: got %d, want 2", len(roles))
	}
}

func TestGetAllCliRoles_Empty(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllCliRoleFn: func() ([]*db_models.CliRole, error) {
			return []*db_models.CliRole{}, nil
		},
	})

	roles, err := service.GetAllCliRoles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("expected empty list, got %d roles", len(roles))
	}
}

func TestGetAllCliRoles_DBError(t *testing.T) {
	dbErr := errors.New("connection refused")
	store.SetSingleton(&testutil.MockStore{
		GetAllCliRoleFn: func() ([]*db_models.CliRole, error) { return nil, dbErr },
	})

	_, err := service.GetAllCliRoles()
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── GetAllUserRolesMappingById ────────────────────────────────────────────────

func TestGetAllUserRolesMappingById_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetRolesByIdFn: func(userID int64) ([]*db_models.CliRoleUserMapping, error) {
			return []*db_models.CliRoleUserMapping{
				{UserID: userID, Permission: "admin"},
			}, nil
		},
	})

	roles, err := service.GetAllUserRolesMappingById(42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 1 || roles[0].Permission != "admin" {
		t.Errorf("unexpected roles: %v", roles)
	}
}

func TestGetAllUserRolesMappingById_DBError(t *testing.T) {
	dbErr := errors.New("query failed")
	store.SetSingleton(&testutil.MockStore{
		GetRolesByIdFn: func(_ int64) ([]*db_models.CliRoleUserMapping, error) {
			return nil, dbErr
		},
	})

	_, err := service.GetAllUserRolesMappingById(1)
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── AddUserRole ───────────────────────────────────────────────────────────────

func TestAddUserRole_Success(t *testing.T) {
	var got *db_models.CliRoleUserMapping
	store.SetSingleton(&testutil.MockStore{
		AddRoleFn: func(role *db_models.CliRoleUserMapping) error {
			got = role
			return nil
		},
	})

	r := &db_models.CliRoleUserMapping{UserID: 1, Permission: "viewer"}
	if err := service.AddUserRole(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != r {
		t.Error("AddUserRole did not pass the correct mapping to store")
	}
}

func TestAddUserRole_DBError(t *testing.T) {
	dbErr := errors.New("insert failed")
	store.SetSingleton(&testutil.MockStore{
		AddRoleFn: func(_ *db_models.CliRoleUserMapping) error { return dbErr },
	})

	if err := service.AddUserRole(&db_models.CliRoleUserMapping{}); !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── DeleteUserRole ────────────────────────────────────────────────────────────

func TestDeleteUserRole_Success(t *testing.T) {
	var deleted *db_models.CliRoleUserMapping
	store.SetSingleton(&testutil.MockStore{
		DeleteRoleFn: func(role *db_models.CliRoleUserMapping) error {
			deleted = role
			return nil
		},
	})

	r := &db_models.CliRoleUserMapping{UserID: 2, Permission: "admin"}
	if err := service.DeleteUserRole(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != r {
		t.Error("DeleteUserRole did not pass the correct mapping to store")
	}
}

func TestDeleteUserRole_DBError(t *testing.T) {
	dbErr := errors.New("delete failed")
	store.SetSingleton(&testutil.MockStore{
		DeleteRoleFn: func(_ *db_models.CliRoleUserMapping) error { return dbErr },
	})

	if err := service.DeleteUserRole(&db_models.CliRoleUserMapping{}); !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}
