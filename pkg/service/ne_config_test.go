package service_test

import (
	"errors"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

// ── CreateNeConfig ────────────────────────────────────────────────────────────

func TestCreateNeConfig_Success(t *testing.T) {
	var got *db_models.CliNeConfig
	store.SetSingleton(&testutil.MockStore{
		CreateCliNeConfigFn: func(cfg *db_models.CliNeConfig) error {
			got = cfg
			return nil
		},
	})

	cfg := &db_models.CliNeConfig{NeID: 1, IPAddress: "10.0.0.1", Port: 22, Protocol: "SSH"}
	if err := service.CreateNeConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != cfg {
		t.Error("CreateNeConfig did not pass the correct config to store")
	}
}

func TestCreateNeConfig_DBError(t *testing.T) {
	dbErr := errors.New("insert failed")
	store.SetSingleton(&testutil.MockStore{
		CreateCliNeConfigFn: func(_ *db_models.CliNeConfig) error { return dbErr },
	})

	err := service.CreateNeConfig(&db_models.CliNeConfig{NeID: 1, IPAddress: "10.0.0.1"})
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── GetNeConfigByNeId ─────────────────────────────────────────────────────────

func TestGetNeConfigByNeId_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeConfigByNeIdFn: func(neId int64) ([]*db_models.CliNeConfig, error) {
			return []*db_models.CliNeConfig{
				{ID: 1, NeID: neId, IPAddress: "10.0.0.1", Protocol: "SSH"},
				{ID: 2, NeID: neId, IPAddress: "10.0.0.2", Protocol: "TELNET"},
			}, nil
		},
	})

	list, err := service.GetNeConfigByNeId(5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("count: got %d, want 2", len(list))
	}
	if list[0].NeID != 5 {
		t.Errorf("NeID: got %d, want 5", list[0].NeID)
	}
}

func TestGetNeConfigByNeId_Empty(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeConfigByNeIdFn: func(_ int64) ([]*db_models.CliNeConfig, error) {
			return []*db_models.CliNeConfig{}, nil
		},
	})

	list, err := service.GetNeConfigByNeId(99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d items", len(list))
	}
}

func TestGetNeConfigByNeId_DBError(t *testing.T) {
	dbErr := errors.New("query failed")
	store.SetSingleton(&testutil.MockStore{
		GetCliNeConfigByNeIdFn: func(_ int64) ([]*db_models.CliNeConfig, error) {
			return nil, dbErr
		},
	})

	_, err := service.GetNeConfigByNeId(1)
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── GetNeConfigById ───────────────────────────────────────────────────────────

func TestGetNeConfigById_Found(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeConfigByIdFn: func(id int64) (*db_models.CliNeConfig, error) {
			return &db_models.CliNeConfig{ID: id, NeID: 3, IPAddress: "192.168.1.1", Protocol: "NETCONF"}, nil
		},
	})

	cfg, err := service.GetNeConfigById(7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ID != 7 {
		t.Errorf("ID: got %d, want 7", cfg.ID)
	}
	if cfg.Protocol != "NETCONF" {
		t.Errorf("Protocol: got %q, want %q", cfg.Protocol, "NETCONF")
	}
}

func TestGetNeConfigById_NotFound(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeConfigByIdFn: func(_ int64) (*db_models.CliNeConfig, error) {
			return nil, nil
		},
	})

	cfg, err := service.GetNeConfigById(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil, got %+v", cfg)
	}
}

func TestGetNeConfigById_DBError(t *testing.T) {
	dbErr := errors.New("connection refused")
	store.SetSingleton(&testutil.MockStore{
		GetCliNeConfigByIdFn: func(_ int64) (*db_models.CliNeConfig, error) {
			return nil, dbErr
		},
	})

	_, err := service.GetNeConfigById(1)
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── UpdateNeConfig ────────────────────────────────────────────────────────────

func TestUpdateNeConfig_Success(t *testing.T) {
	var updated *db_models.CliNeConfig
	store.SetSingleton(&testutil.MockStore{
		UpdateCliNeConfigFn: func(cfg *db_models.CliNeConfig) error {
			updated = cfg
			return nil
		},
	})

	cfg := &db_models.CliNeConfig{ID: 3, NeID: 1, IPAddress: "10.0.0.5", Port: 830, Protocol: "NETCONF"}
	if err := service.UpdateNeConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated != cfg {
		t.Error("UpdateNeConfig did not pass the correct config to store")
	}
}

func TestUpdateNeConfig_DBError(t *testing.T) {
	dbErr := errors.New("update failed")
	store.SetSingleton(&testutil.MockStore{
		UpdateCliNeConfigFn: func(_ *db_models.CliNeConfig) error { return dbErr },
	})

	err := service.UpdateNeConfig(&db_models.CliNeConfig{ID: 1})
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── DeleteNeConfigById ────────────────────────────────────────────────────────

func TestDeleteNeConfigById_Success(t *testing.T) {
	var deletedID int64
	store.SetSingleton(&testutil.MockStore{
		DeleteCliNeConfigByIdFn: func(id int64) error {
			deletedID = id
			return nil
		},
	})

	if err := service.DeleteNeConfigById(42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedID != 42 {
		t.Errorf("deleted ID: got %d, want 42", deletedID)
	}
}

func TestDeleteNeConfigById_DBError(t *testing.T) {
	dbErr := errors.New("delete failed")
	store.SetSingleton(&testutil.MockStore{
		DeleteCliNeConfigByIdFn: func(_ int64) error { return dbErr },
	})

	if err := service.DeleteNeConfigById(1); !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

// ── DeleteNeById (cascade) ────────────────────────────────────────────────────

func TestDeleteNeById_Cascade_AllStepsCalled(t *testing.T) {
	called := map[string]bool{}
	store.SetSingleton(&testutil.MockStore{
		DeleteAllUserNeMappingByNeIdFn: func(neId int64) error { called["user_ne_mapping"] = true; return nil },
		DeleteNeMonitorByNeIdFn:        func(neId int64) error { called["ne_monitor"] = true; return nil },
		DeleteCliNeConfigByNeIdFn:      func(neId int64) error { called["ne_config"] = true; return nil },
		DeleteCliNeSlaveByNeIdFn:       func(neId int64) error { called["ne_slave"] = true; return nil },
		DeleteCliNeByIdFn:              func(id int64) error { called["ne"] = true; return nil },
	})

	if err := service.DeleteNeById(10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, step := range []string{"user_ne_mapping", "ne_monitor", "ne_config", "ne_slave", "ne"} {
		if !called[step] {
			t.Errorf("cascade step %q was not called", step)
		}
	}
}

func TestDeleteNeById_Cascade_StopsOnUserNeMappingError(t *testing.T) {
	dbErr := errors.New("mapping delete failed")
	neMonitorCalled := false
	store.SetSingleton(&testutil.MockStore{
		DeleteAllUserNeMappingByNeIdFn: func(_ int64) error { return dbErr },
		DeleteNeMonitorByNeIdFn:        func(_ int64) error { neMonitorCalled = true; return nil },
		DeleteCliNeConfigByNeIdFn:      func(_ int64) error { return nil },
		DeleteCliNeSlaveByNeIdFn:       func(_ int64) error { return nil },
		DeleteCliNeByIdFn:              func(_ int64) error { return nil },
	})

	err := service.DeleteNeById(5)
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
	if neMonitorCalled {
		t.Error("ne_monitor step should not have been called after user_ne_mapping failure")
	}
}

func TestDeleteNeById_Cascade_StopsOnNeConfigError(t *testing.T) {
	dbErr := errors.New("ne_config delete failed")
	neSlaveCalled := false
	store.SetSingleton(&testutil.MockStore{
		DeleteAllUserNeMappingByNeIdFn: func(_ int64) error { return nil },
		DeleteNeMonitorByNeIdFn:        func(_ int64) error { return nil },
		DeleteCliNeConfigByNeIdFn:      func(_ int64) error { return dbErr },
		DeleteCliNeSlaveByNeIdFn:       func(_ int64) error { neSlaveCalled = true; return nil },
		DeleteCliNeByIdFn:              func(_ int64) error { return nil },
	})

	err := service.DeleteNeById(7)
	if !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
	if neSlaveCalled {
		t.Error("ne_slave step should not have been called after ne_config failure")
	}
}

func TestDeleteNeById_Cascade_StopsOnFinalDeleteError(t *testing.T) {
	dbErr := errors.New("ne delete failed")
	store.SetSingleton(&testutil.MockStore{
		DeleteAllUserNeMappingByNeIdFn: func(_ int64) error { return nil },
		DeleteNeMonitorByNeIdFn:        func(_ int64) error { return nil },
		DeleteCliNeConfigByNeIdFn:      func(_ int64) error { return nil },
		DeleteCliNeSlaveByNeIdFn:       func(_ int64) error { return nil },
		DeleteCliNeByIdFn:              func(_ int64) error { return dbErr },
	})

	if err := service.DeleteNeById(1); !errors.Is(err, dbErr) {
		t.Errorf("error: got %v, want %v", err, dbErr)
	}
}

func TestDeleteNeById_CorrectIdPropagated(t *testing.T) {
	var capturedID int64
	store.SetSingleton(&testutil.MockStore{
		DeleteAllUserNeMappingByNeIdFn: func(id int64) error { capturedID = id; return nil },
		DeleteNeMonitorByNeIdFn:        func(_ int64) error { return nil },
		DeleteCliNeConfigByNeIdFn:      func(_ int64) error { return nil },
		DeleteCliNeSlaveByNeIdFn:       func(_ int64) error { return nil },
		DeleteCliNeByIdFn:              func(_ int64) error { return nil },
	})

	const wantID = int64(999)
	if err := service.DeleteNeById(wantID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedID != wantID {
		t.Errorf("ID: got %d, want %d", capturedID, wantID)
	}
}
