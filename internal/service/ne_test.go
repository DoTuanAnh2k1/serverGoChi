package service_test

import (
	"errors"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/internal/service"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/store"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/testutil"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// ── GetNeListBySystemType ─────────────────────────────────────────────────────

func TestGetNeListBySystemType_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeListBySystemTypeFn: func(systemType string) ([]*db_models.CliNe, error) {
			return []*db_models.CliNe{
				{ID: 1, Name: "NE-01", SystemType: systemType},
				{ID: 2, Name: "NE-02", SystemType: systemType},
			}, nil
		},
	})

	list, err := service.GetNeListBySystemType("router")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("count: got %d, want 2", len(list))
	}
	if list[0].SystemType != "router" {
		t.Errorf("system_type: got %q, want %q", list[0].SystemType, "router")
	}
}

func TestGetNeListBySystemType_Empty(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeListBySystemTypeFn: func(_ string) ([]*db_models.CliNe, error) {
			return []*db_models.CliNe{}, nil
		},
	})

	list, err := service.GetNeListBySystemType("switch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}

func TestGetNeListBySystemType_DBError(t *testing.T) {
	dbErr := errors.New("query failed")
	store.SetSingleton(&testutil.MockStore{
		GetCliNeListBySystemTypeFn: func(_ string) ([]*db_models.CliNe, error) {
			return nil, dbErr
		},
	})

	_, err := service.GetNeListBySystemType("router")
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── GetNeByNeId ───────────────────────────────────────────────────────────────

func TestGetNeByNeId_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeByNeIdFn: func(id int64) (*db_models.CliNe, error) {
			return &db_models.CliNe{ID: id, Name: "NE-HCM-01", IPAddress: "10.0.0.1"}, nil
		},
	})

	ne, err := service.GetNeByNeId(7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ne.ID != 7 {
		t.Errorf("ID: got %d, want 7", ne.ID)
	}
	if ne.Name != "NE-HCM-01" {
		t.Errorf("name: got %q, want %q", ne.Name, "NE-HCM-01")
	}
}

func TestGetNeByNeId_NotFound(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeByNeIdFn: func(_ int64) (*db_models.CliNe, error) {
			return nil, nil
		},
	})

	ne, err := service.GetNeByNeId(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ne != nil {
		t.Errorf("expected nil, got %v", ne)
	}
}

func TestGetNeByNeId_DBError(t *testing.T) {
	dbErr := errors.New("connection refused")
	store.SetSingleton(&testutil.MockStore{
		GetCliNeByNeIdFn: func(_ int64) (*db_models.CliNe, error) {
			return nil, dbErr
		},
	})

	_, err := service.GetNeByNeId(1)
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── GetAllCliNeOfUserByUserId ─────────────────────────────────────────────────

func TestGetAllCliNeOfUserByUserId_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetAllNeOfUserByUserIdFn: func(userID int64) ([]*db_models.CliUserNeMapping, error) {
			return []*db_models.CliUserNeMapping{
				{UserID: userID, TblNeID: 1},
				{UserID: userID, TblNeID: 2},
			}, nil
		},
	})

	mappings, err := service.GetAllCliNeOfUserByUserId(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mappings) != 2 {
		t.Errorf("count: got %d, want 2", len(mappings))
	}
}

func TestGetAllCliNeOfUserByUserId_DBError(t *testing.T) {
	dbErr := errors.New("query failed")
	store.SetSingleton(&testutil.MockStore{
		GetAllNeOfUserByUserIdFn: func(_ int64) ([]*db_models.CliUserNeMapping, error) {
			return nil, dbErr
		},
	})

	_, err := service.GetAllCliNeOfUserByUserId(1)
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── AddUserCliNe ──────────────────────────────────────────────────────────────

func TestAddUserCliNe_Success(t *testing.T) {
	var got *db_models.CliUserNeMapping
	store.SetSingleton(&testutil.MockStore{
		CreateUserNeMappingFn: func(m *db_models.CliUserNeMapping) error {
			got = m
			return nil
		},
	})

	mapping := &db_models.CliUserNeMapping{UserID: 1, TblNeID: 10}
	if err := service.AddUserCliNe(mapping); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != mapping {
		t.Error("AddUserCliNe did not pass the correct mapping to store")
	}
}

func TestAddUserCliNe_DBError(t *testing.T) {
	dbErr := errors.New("insert failed")
	store.SetSingleton(&testutil.MockStore{
		CreateUserNeMappingFn: func(_ *db_models.CliUserNeMapping) error { return dbErr },
	})

	if err := service.AddUserCliNe(&db_models.CliUserNeMapping{}); !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── DeleteCliNe ───────────────────────────────────────────────────────────────

func TestDeleteCliNe_Success(t *testing.T) {
	var deleted *db_models.CliUserNeMapping
	store.SetSingleton(&testutil.MockStore{
		DeleteUserNeMappingFn: func(m *db_models.CliUserNeMapping) error {
			deleted = m
			return nil
		},
	})

	mapping := &db_models.CliUserNeMapping{UserID: 3, TblNeID: 7}
	if err := service.DeleteCliNe(mapping); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != mapping {
		t.Error("DeleteCliNe did not pass the correct mapping to store")
	}
}

func TestDeleteCliNe_DBError(t *testing.T) {
	dbErr := errors.New("delete failed")
	store.SetSingleton(&testutil.MockStore{
		DeleteUserNeMappingFn: func(_ *db_models.CliUserNeMapping) error { return dbErr },
	})

	if err := service.DeleteCliNe(&db_models.CliUserNeMapping{}); !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── GetNeMonitorById ──────────────────────────────────────────────────────────

func TestGetNeMonitorById_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetNeMonitorByIdFn: func(id int64) (*db_models.CliNeMonitor, error) {
			return &db_models.CliNeMonitor{NeID: id, NeName: "NE-HCM-01"}, nil
		},
	})

	monitor, err := service.GetNeMonitorById(99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if monitor.NeID != 99 {
		t.Errorf("NeID: got %d, want 99", monitor.NeID)
	}
	if monitor.NeName != "NE-HCM-01" {
		t.Errorf("NeName: got %q, want %q", monitor.NeName, "NE-HCM-01")
	}
}

func TestGetNeMonitorById_DBError(t *testing.T) {
	dbErr := errors.New("query failed")
	store.SetSingleton(&testutil.MockStore{
		GetNeMonitorByIdFn: func(_ int64) (*db_models.CliNeMonitor, error) {
			return nil, dbErr
		},
	})

	_, err := service.GetNeMonitorById(1)
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}
