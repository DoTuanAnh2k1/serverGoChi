package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/config"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/testutil"
	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
)

func TestMain(m *testing.M) {
	testutil.InitTestLogger()
	config.Init(&config_models.Config{})
	os.Exit(m.Run())
}

func decode(t *testing.T, w *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(w.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

// ── Write ─────────────────────────────────────────────────────────────────────

func TestWrite_SetsContentTypeAndCode(t *testing.T) {
	w := httptest.NewRecorder()
	response.Write(w, http.StatusTeapot, map[string]string{"k": "v"})

	if w.Code != http.StatusTeapot {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusTeapot)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type: got %q, want %q", ct, "application/json")
	}
}

// ── Success ───────────────────────────────────────────────────────────────────

func TestSuccess_DefaultMessage(t *testing.T) {
	w := httptest.NewRecorder()
	response.Success(w, "")

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	var r response.ResSuccess
	decode(t, w, &r)
	if r.Message != "Success" {
		t.Errorf("message: got %q, want %q", r.Message, "Success")
	}
	if !r.Status {
		t.Error("status field should be true")
	}
}

func TestSuccess_CustomMessage(t *testing.T) {
	w := httptest.NewRecorder()
	response.Success(w, "operation complete")

	var r response.ResSuccess
	decode(t, w, &r)
	if r.Message != "operation complete" {
		t.Errorf("message: got %q, want %q", r.Message, "operation complete")
	}
}

// ── Created ───────────────────────────────────────────────────────────────────

func TestCreated(t *testing.T) {
	w := httptest.NewRecorder()
	response.Created(w)

	if w.Code != http.StatusCreated {
		t.Errorf("status: got %d, want 201", w.Code)
	}
	var r response.ResSuccess
	decode(t, w, &r)
	if r.Message != "Created" {
		t.Errorf("message: got %q, want %q", r.Message, "Created")
	}
}

// ── Updated ───────────────────────────────────────────────────────────────────

func TestUpdated(t *testing.T) {
	w := httptest.NewRecorder()
	response.Updated(w)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", w.Code)
	}
	var r response.ResSuccess
	decode(t, w, &r)
	if r.Message != "Updated" {
		t.Errorf("message: got %q, want %q", r.Message, "Updated")
	}
}

// ── NoContent ─────────────────────────────────────────────────────────────────

func TestNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	response.NoContent(w)

	if w.Code != http.StatusNoContent {
		t.Errorf("status: got %d, want 204", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("body should be empty, got %q", w.Body.String())
	}
}

// ── NotFound ──────────────────────────────────────────────────────────────────

func TestNotFound_DefaultMessage(t *testing.T) {
	w := httptest.NewRecorder()
	response.NotFound(w, "")

	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want 404", w.Code)
	}
	var r response.ResError
	decode(t, w, &r)
	if r.Status {
		t.Error("status field should be false")
	}
}

func TestNotFound_CustomMessage(t *testing.T) {
	w := httptest.NewRecorder()
	response.NotFound(w, "resource not found")

	var r response.ResError
	decode(t, w, &r)
	if r.Error != "resource not found" {
		t.Errorf("error field: got %q, want %q", r.Error, "resource not found")
	}
}

// ── Unauthorized ──────────────────────────────────────────────────────────────

func TestUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	response.Unauthorized(w)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", w.Code)
	}
	var r response.ResError
	decode(t, w, &r)
	if r.Status {
		t.Error("status field should be false")
	}
	if r.Message != "Unauthorized" {
		t.Errorf("message: got %q, want %q", r.Message, "Unauthorized")
	}
}

// ── Authenticate ─────────────────────────────────────────────────────────────

func TestAuthenticate_SetsWWWAuthenticateHeader(t *testing.T) {
	w := httptest.NewRecorder()
	response.Authenticate(w)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", w.Code)
	}
	if h := w.Header().Get("WWW-Authenticate"); h == "" {
		t.Error("WWW-Authenticate header should be set")
	}
}

// ── InternalError ─────────────────────────────────────────────────────────────

func TestInternalError_DefaultMessage(t *testing.T) {
	w := httptest.NewRecorder()
	response.InternalError(w, "")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want 500", w.Code)
	}
}

func TestInternalError_CustomMessage(t *testing.T) {
	w := httptest.NewRecorder()
	response.InternalError(w, "db connection lost")

	var r response.ResError
	decode(t, w, &r)
	if r.Error != "db connection lost" {
		t.Errorf("error field: got %q, want %q", r.Error, "db connection lost")
	}
}

// ── BadRequest ────────────────────────────────────────────────────────────────

func TestBadRequest_DefaultMessage(t *testing.T) {
	w := httptest.NewRecorder()
	response.BadRequest(w, "")

	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", w.Code)
	}
}

func TestBadRequest_CustomMessage(t *testing.T) {
	w := httptest.NewRecorder()
	response.BadRequest(w, "missing field")

	var r response.ResError
	decode(t, w, &r)
	if r.Error != "missing field" {
		t.Errorf("error field: got %q, want %q", r.Error, "missing field")
	}
}

// ── BadGateway ────────────────────────────────────────────────────────────────

func TestBadGateway(t *testing.T) {
	w := httptest.NewRecorder()
	response.BadGateway(w, "upstream unavailable")

	if w.Code != http.StatusBadGateway {
		t.Errorf("status: got %d, want 502", w.Code)
	}
	var r response.ResError
	decode(t, w, &r)
	if r.Error != "upstream unavailable" {
		t.Errorf("error field: got %q, want %q", r.Error, "upstream unavailable")
	}
}

// ── MethodNotAllowed ──────────────────────────────────────────────────────────

func TestMethodNotAllowed_DefaultMessage(t *testing.T) {
	w := httptest.NewRecorder()
	response.MethodNotAllowed(w, "")

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status: got %d, want 405", w.Code)
	}
}

func TestMethodNotAllowed_CustomMessage(t *testing.T) {
	w := httptest.NewRecorder()
	response.MethodNotAllowed(w, "only GET allowed")

	var r response.ResError
	decode(t, w, &r)
	if r.Error != "only GET allowed" {
		t.Errorf("error field: got %q, want %q", r.Error, "only GET allowed")
	}
}
