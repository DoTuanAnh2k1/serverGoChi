package sshcli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeJWT returns a JWT-shaped token with the given "aud" claim. The signature
// isn't verified by our code (we trust the upstream that issued it), so any
// string in the third segment is fine.
func fakeJWT(aud string) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"aud":%q,"sub":"alice","exp":9999999999}`, aud)))
	return "Basic " + header + "." + payload + ".sig"
}

// fakeServer returns a handler that dispatches by method+path to the provided map.
type route struct {
	method string
	path   string
}

func newFakeMgt(t *testing.T, handlers map[route]http.HandlerFunc) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	seen := map[string]bool{}
	for k := range handlers {
		k := k
		if seen[k.path] {
			continue
		}
		seen[k.path] = true
		mux.HandleFunc(k.path, func(w http.ResponseWriter, r *http.Request) {
			h, ok := handlers[route{method: r.Method, path: r.URL.Path}]
			if !ok {
				t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
				http.Error(w, "unexpected", http.StatusBadRequest)
				return
			}
			h(w, r)
		})
	}
	return httptest.NewServer(mux)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func readJSON(t *testing.T, r *http.Request, v any) {
	t.Helper()
	body, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(body, v); err != nil {
		t.Fatalf("decode request body: %v (%s)", err, body)
	}
}

func TestAuthenticate_Success(t *testing.T) {
	adminTok := fakeJWT("admin")
	srv := newFakeMgt(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/authenticate"}: func(w http.ResponseWriter, r *http.Request) {
			var body map[string]string
			readJSON(t, r, &body)
			if body["username"] != "alice" || body["password"] != "pw" {
				t.Errorf("wrong creds: %+v", body)
			}
			writeJSON(w, 200, map[string]any{
				"status":        "success",
				"response_data": adminTok,
				"response_code": "200",
				"system_type":   "5GC",
			})
		},
	})
	defer srv.Close()

	c := NewMgtClient(srv.URL)
	if err := c.Authenticate("alice", "pw"); err != nil {
		t.Fatalf("err: %v", err)
	}
	if c.Token != adminTok {
		t.Errorf("token: %q", c.Token)
	}
	if c.Role != "admin" {
		t.Errorf("role: %q", c.Role)
	}
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	srv := newFakeMgt(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/authenticate"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 401, map[string]string{"status": "error"})
		},
	})
	defer srv.Close()

	c := NewMgtClient(srv.URL)
	err := c.Authenticate("alice", "bad")
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected auth error with 401, got %v", err)
	}
}

// Normal users (role=="user") are now admitted at the auth layer. The mode
// menu filters cli-config out for them; the downstream ne-config / ne-command
// services enforce per-command authorization via /aa/authorize/rbac.
func TestAuthenticate_AcceptsNormalUser(t *testing.T) {
	userTok := fakeJWT("user")
	srv := newFakeMgt(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/authenticate"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, map[string]any{
				"status":        "success",
				"response_data": userTok,
				"response_code": "200",
			})
		},
	})
	defer srv.Close()

	c := NewMgtClient(srv.URL)
	if err := c.Authenticate("alice", "pw"); err != nil {
		t.Fatalf("normal user should authenticate: %v", err)
	}
	if c.Role != "user" {
		t.Errorf("Role: got %q, want %q", c.Role, "user")
	}
}

func TestCreateUser(t *testing.T) {
	var gotBody map[string]any
	srv := newFakeMgt(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/authenticate/user/set"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &gotBody)
			if r.Header.Get("Authorization") != "Basic tok" {
				t.Errorf("auth: %q", r.Header.Get("Authorization"))
			}
			writeJSON(w, 201, map[string]string{"status": "success"})
		},
	})
	defer srv.Close()

	c := NewMgtClient(srv.URL)
	c.Token = "Basic tok"
	if err := c.CreateUser(map[string]any{
		"account_name": "alice",
		"password":     "pw",
		"account_type": 2,
	}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if gotBody["account_name"] != "alice" {
		t.Errorf("body: %+v", gotBody)
	}
	// JSON decodes int as float64
	if gotBody["account_type"] != float64(2) {
		t.Errorf("account_type: %v", gotBody["account_type"])
	}
}

func TestCreateUser_ServerError(t *testing.T) {
	srv := newFakeMgt(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/authenticate/user/set"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 400, map[string]string{"status": "error", "message": "email invalid"})
		},
	})
	defer srv.Close()

	c := NewMgtClient(srv.URL)
	c.Token = "Basic tok"
	err := c.CreateUser(map[string]any{"account_name": "a", "password": "p"})
	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Errorf("expected 400 error, got %v", err)
	}
}

func TestListNEs(t *testing.T) {
	srv := newFakeMgt(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/admin/ne/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []NeInfo{{ID: 10, NeName: "HTSMF01", Namespace: "hcm"}})
		},
	})
	defer srv.Close()

	c := NewMgtClient(srv.URL)
	c.Token = "Basic tok"
	nes, err := c.ListNEs()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(nes) != 1 || nes[0].NeName != "HTSMF01" {
		t.Errorf("nes: %+v", nes)
	}
}

func TestResolveNEID(t *testing.T) {
	srv := newFakeMgt(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/admin/ne/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []NeInfo{{ID: 10, NeName: "HTSMF01"}, {ID: 20, NeName: "HTSMF02"}})
		},
	})
	defer srv.Close()

	c := NewMgtClient(srv.URL)
	c.Token = "Basic tok"
	id, err := c.ResolveNEID("42")
	if err != nil || id != 42 {
		t.Errorf("numeric: id=%d err=%v", id, err)
	}
	id, err = c.ResolveNEID("HTSMF02")
	if err != nil || id != 20 {
		t.Errorf("by name: id=%d err=%v", id, err)
	}
	_, err = c.ResolveNEID("NONEXISTENT")
	if err == nil {
		t.Errorf("missing name should error")
	}
}

func TestAssignNEToUser(t *testing.T) {
	var gotBody map[string]string
	srv := newFakeMgt(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/authorize/ne/set"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &gotBody)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer srv.Close()

	c := NewMgtClient(srv.URL)
	c.Token = "Basic tok"
	if err := c.AssignNEToUser("alice", 42); err != nil {
		t.Fatalf("err: %v", err)
	}
	if gotBody["username"] != "alice" || gotBody["neid"] != "42" {
		t.Errorf("body: %+v", gotBody)
	}
}

func TestGroupCRUD(t *testing.T) {
	var createdBody map[string]any
	srv := newFakeMgt(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/group/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []GroupInfo{{ID: 5, Name: "dev"}})
		},
		{http.MethodPost, "/aa/group/create"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &createdBody)
			writeJSON(w, 201, map[string]string{"status": "success"})
		},
		{http.MethodPost, "/aa/group/ne/assign"}: func(w http.ResponseWriter, r *http.Request) {
			var b map[string]int64
			readJSON(t, r, &b)
			if b["group_id"] != 5 || b["ne_id"] != 10 {
				t.Errorf("body: %+v", b)
			}
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer srv.Close()

	c := NewMgtClient(srv.URL)
	c.Token = "Basic tok"

	gs, err := c.ListGroups()
	if err != nil || len(gs) != 1 || gs[0].Name != "dev" {
		t.Fatalf("groups: %+v err=%v", gs, err)
	}
	if err := c.CreateGroup(map[string]any{"name": "ops", "description": "ops team"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if createdBody["name"] != "ops" {
		t.Errorf("created: %+v", createdBody)
	}
	if err := c.AssignNEToGroup(5, 10); err != nil {
		t.Fatalf("assign: %v", err)
	}
	id, err := c.ResolveGroupID("dev")
	if err != nil || id != 5 {
		t.Errorf("resolve group: id=%d err=%v", id, err)
	}
}

func TestDeleteUser(t *testing.T) {
	var gotBody map[string]string
	srv := newFakeMgt(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/authenticate/user/delete"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &gotBody)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer srv.Close()

	c := NewMgtClient(srv.URL)
	c.Token = "Basic tok"
	if err := c.DeleteUser("alice"); err != nil {
		t.Fatalf("err: %v", err)
	}
	if gotBody["account_name"] != "alice" {
		t.Errorf("body: %+v", gotBody)
	}
}
