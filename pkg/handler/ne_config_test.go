package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
)

// helpers

func neConfigBody(t *testing.T, v interface{}) *bytes.Buffer {
	t.Helper()
	b, _ := json.Marshal(v)
	return bytes.NewBuffer(b)
}

func adminUser() *middleware.User { return &middleware.User{Username: "admin", Roles: "admin"} }

// ── HandlerNeConfigCreate ─────────────────────────────────────────────────────

func TestHandlerNeConfigCreate_Success(t *testing.T) {
	var saved db_models.CliNeConfig
	store.SetSingleton(&testutil.MockStore{
		CreateCliNeConfigFn: func(cfg *db_models.CliNeConfig) error {
			saved = *cfg
			return nil
		},
		SaveHistoryCommandFn: func(_ db_models.CliOperationHistory) error { return nil },
	})

	body := neConfigBody(t, map[string]interface{}{
		"ne_id": 1, "ip_address": "10.0.0.1", "port": 22, "protocol": "SSH",
	})
	req := httptest.NewRequest(http.MethodPost, "/aa/authorize/ne/config/create", body)
	req.Header.Set("Content-Type", "application/json")
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigCreate(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status: got %d, want 201", w.Code)
	}
	if saved.IPAddress != "10.0.0.1" {
		t.Errorf("IPAddress: got %q, want %q", saved.IPAddress, "10.0.0.1")
	}
	if saved.NeID != 1 {
		t.Errorf("NeID: got %d, want 1", saved.NeID)
	}
}

func TestHandlerNeConfigCreate_DefaultProtocol(t *testing.T) {
	var saved db_models.CliNeConfig
	store.SetSingleton(&testutil.MockStore{
		CreateCliNeConfigFn: func(cfg *db_models.CliNeConfig) error { saved = *cfg; return nil },
		SaveHistoryCommandFn: func(_ db_models.CliOperationHistory) error { return nil },
	})

	body := neConfigBody(t, map[string]interface{}{"ne_id": 2, "ip_address": "10.0.0.2"})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigCreate(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("status: got %d, want 201", w.Code)
	}
	if saved.Protocol != "SSH" {
		t.Errorf("default Protocol: got %q, want %q", saved.Protocol, "SSH")
	}
}

func TestHandlerNeConfigCreate_MissingNeId(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	body := neConfigBody(t, map[string]interface{}{"ip_address": "10.0.0.1"})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigCreate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerNeConfigCreate_MissingIPAddress(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	body := neConfigBody(t, map[string]interface{}{"ne_id": 1})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigCreate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerNeConfigCreate_DBError(t *testing.T) {
	dbErr := errors.New("insert failed")
	store.SetSingleton(&testutil.MockStore{
		CreateCliNeConfigFn:  func(_ *db_models.CliNeConfig) error { return dbErr },
		SaveHistoryCommandFn: func(_ db_models.CliOperationHistory) error { return nil },
	})

	body := neConfigBody(t, map[string]interface{}{"ne_id": 1, "ip_address": "10.0.0.1"})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigCreate(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

// ── HandlerNeConfigList ───────────────────────────────────────────────────────

func TestHandlerNeConfigList_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeConfigByNeIdFn: func(neId int64) ([]*db_models.CliNeConfig, error) {
			return []*db_models.CliNeConfig{
				{ID: 1, NeID: neId, IPAddress: "10.0.0.1", Protocol: "SSH"},
				{ID: 2, NeID: neId, IPAddress: "10.0.0.2", Protocol: "TELNET"},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/aa/authorize/ne/config/list?ne_id=5", nil)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigList(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}

	var list []*db_models.CliNeConfig
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("count: got %d, want 2", len(list))
	}
}

func TestHandlerNeConfigList_EmptyResult(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeConfigByNeIdFn: func(_ int64) ([]*db_models.CliNeConfig, error) {
			return nil, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/?ne_id=1", nil)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigList(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}

	// should return empty array, not null
	body := w.Body.String()
	if body == "null\n" || body == "null" {
		t.Error("expected empty array [], got null")
	}
}

func TestHandlerNeConfigList_MissingNeId(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	req := httptest.NewRequest(http.MethodGet, "/aa/authorize/ne/config/list", nil)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigList(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerNeConfigList_InvalidNeId(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	req := httptest.NewRequest(http.MethodGet, "/?ne_id=abc", nil)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigList(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerNeConfigList_DBError(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeConfigByNeIdFn: func(_ int64) ([]*db_models.CliNeConfig, error) {
			return nil, errors.New("query failed")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/?ne_id=1", nil)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigList(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

// ── HandlerNeConfigUpdate ─────────────────────────────────────────────────────

func TestHandlerNeConfigUpdate_Success(t *testing.T) {
	var updated db_models.CliNeConfig
	store.SetSingleton(&testutil.MockStore{
		UpdateCliNeConfigFn:  func(cfg *db_models.CliNeConfig) error { updated = *cfg; return nil },
		SaveHistoryCommandFn: func(_ db_models.CliOperationHistory) error { return nil },
	})

	body := neConfigBody(t, map[string]interface{}{
		"id": 7, "ne_id": 1, "ip_address": "10.0.0.9", "port": 830, "protocol": "NETCONF",
	})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigUpdate(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	if updated.ID != 7 {
		t.Errorf("ID: got %d, want 7", updated.ID)
	}
	if updated.Protocol != "NETCONF" {
		t.Errorf("Protocol: got %q, want %q", updated.Protocol, "NETCONF")
	}
}

func TestHandlerNeConfigUpdate_MissingId(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	body := neConfigBody(t, map[string]interface{}{"ip_address": "10.0.0.1"})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigUpdate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerNeConfigUpdate_DBError(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		UpdateCliNeConfigFn:  func(_ *db_models.CliNeConfig) error { return errors.New("update failed") },
		SaveHistoryCommandFn: func(_ db_models.CliOperationHistory) error { return nil },
	})

	body := neConfigBody(t, map[string]interface{}{"id": 3, "ip_address": "10.0.0.1"})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigUpdate(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

// ── HandlerNeConfigDelete ─────────────────────────────────────────────────────

func TestHandlerNeConfigDelete_Success(t *testing.T) {
	var deletedID int64
	store.SetSingleton(&testutil.MockStore{
		DeleteCliNeConfigByIdFn: func(id int64) error { deletedID = id; return nil },
		SaveHistoryCommandFn:    func(_ db_models.CliOperationHistory) error { return nil },
	})

	body := neConfigBody(t, map[string]interface{}{"id": 42})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigDelete(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	if deletedID != 42 {
		t.Errorf("deleted ID: got %d, want 42", deletedID)
	}
}

func TestHandlerNeConfigDelete_MissingId(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	body := neConfigBody(t, map[string]interface{}{})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigDelete(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerNeConfigDelete_DBError(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		DeleteCliNeConfigByIdFn: func(_ int64) error { return errors.New("delete failed") },
		SaveHistoryCommandFn:    func(_ db_models.CliOperationHistory) error { return nil },
	})

	body := neConfigBody(t, map[string]interface{}{"id": 1})
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerNeConfigDelete(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

// ── HandlerListNeConfig ───────────────────────────────────────────────────────

func TestHandlerListNeConfig_Success(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(_ string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 1, AccountName: "admin"}, nil
		},
		GetAllNeOfUserByUserIdFn: func(_ int64) ([]*db_models.CliUserNeMapping, error) {
			return []*db_models.CliUserNeMapping{
				{UserID: 1, TblNeID: 10},
			}, nil
		},
		GetCliNeByNeIdFn: func(id int64) (*db_models.CliNe, error) {
			return &db_models.CliNe{ID: id, Name: "NE-HCM-01", IPAddress: "10.0.0.1", SiteName: "HCM"}, nil
		},
		GetCliNeConfigByNeIdFn: func(_ int64) ([]*db_models.CliNeConfig, error) {
			return []*db_models.CliNeConfig{
				{ID: 1, NeID: 10, IPAddress: "10.0.0.1", Protocol: "SSH"},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/aa/list/ne/config", nil)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerListNeConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("NE count: got %d, want 1", len(result))
	}
	if result[0]["ne_name"] != "NE-HCM-01" {
		t.Errorf("ne_name: got %v", result[0]["ne_name"])
	}
	cfgList, ok := result[0]["config_list"].([]interface{})
	if !ok || len(cfgList) != 1 {
		t.Errorf("config_list: got %v", result[0]["config_list"])
	}
}

func TestHandlerListNeConfig_NoUser(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	req := httptest.NewRequest(http.MethodGet, "/aa/list/ne/config", nil)
	// no user injected
	w := httptest.NewRecorder()

	handler.HandlerListNeConfig(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

func TestHandlerListNeConfig_GetUserError(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(_ string) (*db_models.TblAccount, error) {
			return nil, errors.New("db error")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerListNeConfig(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

func TestHandlerListNeConfig_EmptyMappings(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetUserByUserNameFn: func(_ string) (*db_models.TblAccount, error) {
			return &db_models.TblAccount{AccountID: 1}, nil
		},
		GetAllNeOfUserByUserIdFn: func(_ int64) ([]*db_models.CliUserNeMapping, error) {
			return []*db_models.CliUserNeMapping{}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = reqWithUser(req, adminUser())
	w := httptest.NewRecorder()

	handler.HandlerListNeConfig(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
}
