package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// ── HandlerAdminNeUpdate ──────────────────────────────────────────────────────

func TestHandlerAdminNeUpdate_Success(t *testing.T) {
	var saved *db_models.CliNe
	store.SetSingleton(&testutil.MockStore{
		UpdateCliNeFn: func(ne *db_models.CliNe) error {
			saved = ne
			return nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]any{
		"id":                   42,
		"ne_name":              "HTSMF99",
		"namespace":            "hcm-5gc",
		"conf_master_ip":       "10.10.9.9",
		"conf_port_master_tcp": 830,
		"command_url":          "http://10.10.9.9:8080",
		"description":          "edited",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminNeUpdate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	if saved == nil {
		t.Fatal("UpdateCliNe was not called")
	}
	if saved.ID != 42 || saved.NeName != "HTSMF99" || saved.Description != "edited" {
		t.Errorf("saved NE wrong: %+v", saved)
	}
}

func TestHandlerAdminNeUpdate_MissingID(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	body, _ := json.Marshal(map[string]any{"ne_name": "noid"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminNeUpdate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerAdminNeUpdate_InvalidJSON(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("not-json")))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminNeUpdate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestHandlerAdminNeUpdate_DBError(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		UpdateCliNeFn:        func(ne *db_models.CliNe) error { return errors.New("db down") },
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]any{"id": 1})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminNeUpdate(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

// ── HandlerAdminNeCreate ──────────────────────────────────────────────────────

func TestHandlerAdminNeCreate_Success(t *testing.T) {
	var created *db_models.CliNe
	store.SetSingleton(&testutil.MockStore{
		CreateCliNeFn: func(ne *db_models.CliNe) error {
			created = ne
			return nil
		},
		SaveHistoryCommandFn: func(h db_models.CliOperationHistory) error { return nil },
	})

	body, _ := json.Marshal(map[string]any{
		"ne_name":              "HTSMF01",
		"namespace":            "hcm-5gc",
		"conf_master_ip":       "10.10.1.1",
		"conf_port_master_tcp": 830,
		"command_url":          "http://10.10.1.1:8080",
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminNeCreate(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status: got %d, want 201", w.Code)
	}
	if created == nil || created.NeName != "HTSMF01" || created.SystemType != "5GC" {
		t.Errorf("created NE wrong: %+v", created)
	}
}

func TestHandlerAdminNeCreate_MissingFields(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{})

	body, _ := json.Marshal(map[string]any{"ne_name": "only-name"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminNeCreate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

// ── HandlerAdminNeList ────────────────────────────────────────────────────────

func TestHandlerAdminNeList_ReturnsList(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeListBySystemTypeFn: func(systemType string) ([]*db_models.CliNe, error) {
			if systemType != "5GC" {
				t.Errorf("system_type: got %q, want %q", systemType, "5GC")
			}
			return []*db_models.CliNe{
				{ID: 1, NeName: "HTSMF01", Namespace: "hcm-5gc"},
				{ID: 2, NeName: "HTAMF01", Namespace: "hcm-5gc"},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminNeList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
	var got []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 NEs, got %d", len(got))
	}
}

func TestHandlerAdminNeList_Empty(t *testing.T) {
	store.SetSingleton(&testutil.MockStore{
		GetCliNeListBySystemTypeFn: func(systemType string) ([]*db_models.CliNe, error) { return nil, nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = injectUser(req, "alice")
	w := httptest.NewRecorder()

	handler.HandlerAdminNeList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", w.Code)
	}
}
