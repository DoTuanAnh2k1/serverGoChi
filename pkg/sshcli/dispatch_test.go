package sshcli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func newDispatcher(t *testing.T, handlers map[route]http.HandlerFunc) (*Dispatcher, *bytes.Buffer, func()) {
	t.Helper()
	srv := newFakeMgt(t, handlers)
	c := NewMgtClient(srv.URL)
	c.Token = "Basic tok"
	buf := &bytes.Buffer{}
	return &Dispatcher{Client: c, Out: buf}, buf, srv.Close
}

func mustParse(t *testing.T, line string) *Command {
	t.Helper()
	c, err := Parse(line)
	if err != nil {
		t.Fatalf("parse %q: %v", line, err)
	}
	if c == nil {
		t.Fatalf("parse %q: nil", line)
	}
	return c
}

func TestDispatch_ShowUserTable(t *testing.T) {
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/admin/user/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []UserInfo{
				{AccountID: 1, AccountName: "alice", AccountType: 1, IsEnable: true},
				{AccountID: 2, AccountName: "bob", AccountType: 2, IsEnable: true},
			})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "show user")); err != nil {
		t.Fatalf("err: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "alice") || !strings.Contains(out, "Admin") {
		t.Errorf("output missing names: %s", out)
	}
	if !strings.Contains(out, "(2 users)") {
		t.Errorf("count missing: %s", out)
	}
}

func TestDispatch_ShowUserByName(t *testing.T) {
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/admin/user/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []UserInfo{{AccountID: 1, AccountName: "alice", AccountType: 1}})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "show user alice")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(buf.String(), "name:         alice") {
		t.Errorf("detail missing: %s", buf.String())
	}

	if err := d.Run(mustParse(t, "show user nonexistent")); err == nil {
		t.Errorf("expected not-found error")
	}
}

func TestDispatch_SetUserRequiredCheck(t *testing.T) {
	var gotBody map[string]any
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/authenticate/user/set"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &gotBody)
			writeJSON(w, 201, map[string]string{"status": "success"})
		},
	})
	defer done()

	// Missing password → client-side validation catches it.
	if err := d.Run(mustParse(t, "set user name alice email a@b.c")); err == nil {
		t.Errorf("expected required-field error")
	}

	// Happy path.
	if err := d.Run(mustParse(t, "set user name alice password pw email a@b.c")); err != nil {
		t.Fatalf("happy set: %v", err)
	}
	if gotBody["account_name"] != "alice" || gotBody["password"] != "pw" || gotBody["email"] != "a@b.c" {
		t.Errorf("body: %+v", gotBody)
	}
}

func TestDispatch_UpdateNeByName(t *testing.T) {
	var updateBody map[string]any
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/admin/ne/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []NeInfo{{ID: 42, NeName: "HTSMF01"}})
		},
		{http.MethodPost, "/aa/admin/ne/update"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &updateBody)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "update ne HTSMF01 site_name HN")); err != nil {
		t.Fatalf("err: %v", err)
	}
	// id resolved via ListNEs; site_name carried through.
	if updateBody["id"] != float64(42) || updateBody["site_name"] != "HN" {
		t.Errorf("body: %+v", updateBody)
	}
}

func TestDispatch_DeleteNe(t *testing.T) {
	var delBody map[string]any
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/admin/ne/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []NeInfo{{ID: 42, NeName: "HTSMF01"}})
		},
		{http.MethodPost, "/aa/authorize/ne/remove"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &delBody)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "delete ne HTSMF01")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if delBody["id"] != float64(42) {
		t.Errorf("body: %+v", delBody)
	}
}

func TestDispatch_MapUserNe(t *testing.T) {
	var assignBody map[string]string
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/admin/ne/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []NeInfo{{ID: 10, NeName: "HTSMF01"}})
		},
		{http.MethodPost, "/aa/authorize/ne/set"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &assignBody)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "map user alice ne HTSMF01")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if assignBody["username"] != "alice" || assignBody["neid"] != "10" {
		t.Errorf("body: %+v", assignBody)
	}
}

func TestDispatch_UnmapGroupNe(t *testing.T) {
	var body map[string]int64
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/group/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []GroupInfo{{ID: 5, Name: "dev"}})
		},
		{http.MethodGet, "/aa/admin/ne/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []NeInfo{{ID: 10, NeName: "HTSMF01"}})
		},
		{http.MethodPost, "/aa/group/ne/unassign"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &body)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "unmap group dev ne HTSMF01")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if body["group_id"] != 5 || body["ne_id"] != 10 {
		t.Errorf("body: %+v", body)
	}
}

func TestDispatch_ServerError(t *testing.T) {
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/admin/user/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 500, map[string]string{"status": "error"})
		},
	})
	defer done()

	err := d.Run(mustParse(t, "show user"))
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 error, got %v", err)
	}
}

func TestDispatch_Help(t *testing.T) {
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{})
	defer done()

	if err := d.Run(mustParse(t, "help")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(buf.String(), "Available commands") {
		t.Errorf("help general: %s", buf.String())
	}
	buf.Reset()
	if err := d.Run(mustParse(t, "help set")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if !strings.Contains(buf.String(), "Create a new record") {
		t.Errorf("help set: %s", buf.String())
	}
}

func TestDispatch_Exit(t *testing.T) {
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{})
	defer done()
	err := d.Run(mustParse(t, "exit"))
	if err != io.EOF {
		t.Errorf("exit: got %v, want io.EOF", err)
	}
}

// --- show ne ---

func TestDispatch_ShowNeTable(t *testing.T) {
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/admin/ne/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []NeInfo{
				{ID: 10, NeName: "HTSMF01", Namespace: "hcm", ConfMode: "SSH"},
				{ID: 20, NeName: "HTSMF02", Namespace: "hn", ConfMode: "NETCONF"},
			})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "show ne")); err != nil {
		t.Fatalf("err: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"HTSMF01", "HTSMF02", "SSH", "NETCONF", "(2 NEs)"} {
		if !strings.Contains(out, want) {
			t.Errorf("show ne table missing %q:\n%s", want, out)
		}
	}
}

func TestDispatch_ShowNeByName(t *testing.T) {
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/admin/ne/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []NeInfo{{ID: 10, NeName: "HTSMF01", ConfMasterIP: "10.0.0.1", ConfPortMasterTCP: 830, ConfMode: "NETCONF"}})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "show ne HTSMF01")); err != nil {
		t.Fatalf("err: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"ne_name:", "HTSMF01", "conf_master_ip:", "10.0.0.1", "conf_mode:", "NETCONF"} {
		if !strings.Contains(out, want) {
			t.Errorf("show ne detail missing %q:\n%s", want, out)
		}
	}

	if err := d.Run(mustParse(t, "show ne MISSING")); err == nil {
		t.Errorf("expected not-found error")
	}
}

// --- show group list ---

func TestDispatch_ShowGroupTable(t *testing.T) {
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/group/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []GroupInfo{
				{ID: 1, Name: "dev", Description: "developers"},
				{ID: 2, Name: "ops", Description: "operations"},
			})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "show group")); err != nil {
		t.Fatalf("err: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"dev", "ops", "developers", "(2 groups)"} {
		if !strings.Contains(out, want) {
			t.Errorf("show group table missing %q:\n%s", want, out)
		}
	}
}

// --- set ne / set group ---

func TestDispatch_SetNe(t *testing.T) {
	var body map[string]any
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/admin/ne/create"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &body)
			writeJSON(w, 201, map[string]string{"status": "success"})
		},
	})
	defer done()

	// Missing required field (namespace) — should fail client-side.
	if err := d.Run(mustParse(t, "set ne name X conf_master_ip 1.2.3.4 port 830 command_url http://x")); err == nil {
		t.Errorf("expected required-field error")
	}

	// Happy path.
	cmd := "set ne name HTSMF01 namespace hcm conf_master_ip 1.2.3.4 port 830 command_url http://x mode SSH"
	if err := d.Run(mustParse(t, cmd)); err != nil {
		t.Fatalf("happy set: %v", err)
	}
	if body["ne_name"] != "HTSMF01" || body["namespace"] != "hcm" {
		t.Errorf("body: %+v", body)
	}
	if body["conf_port_master_tcp"] != float64(830) {
		t.Errorf("int field: %+v", body["conf_port_master_tcp"])
	}
	if body["conf_mode"] != "SSH" {
		t.Errorf("enum: %+v", body["conf_mode"])
	}
	if !strings.Contains(buf.String(), "OK: NE created") {
		t.Errorf("ack: %s", buf.String())
	}
}

func TestDispatch_SetNe_InvalidEnum(t *testing.T) {
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{})
	defer done()
	cmd := "set ne name X namespace n conf_master_ip 1.1.1.1 port 830 command_url http://x mode BOGUS"
	if err := d.Run(mustParse(t, cmd)); err == nil || !strings.Contains(err.Error(), "must be one of") {
		t.Errorf("expected enum error, got %v", err)
	}
}

func TestDispatch_SetNe_IntParseError(t *testing.T) {
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{})
	defer done()
	cmd := "set ne name X namespace n conf_master_ip 1.1.1.1 port notanumber command_url http://x"
	if err := d.Run(mustParse(t, cmd)); err == nil || !strings.Contains(err.Error(), "integer") {
		t.Errorf("expected int parse error, got %v", err)
	}
}

func TestDispatch_SetGroup(t *testing.T) {
	var body map[string]any
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/group/create"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &body)
			writeJSON(w, 201, map[string]string{"status": "success"})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, `set group name ops description "operations team"`)); err != nil {
		t.Fatalf("err: %v", err)
	}
	if body["name"] != "ops" || body["description"] != "operations team" {
		t.Errorf("body: %+v", body)
	}
	if !strings.Contains(buf.String(), "OK: group created") {
		t.Errorf("ack: %s", buf.String())
	}
}

// --- update user / update group ---

func TestDispatch_UpdateUser(t *testing.T) {
	var body map[string]any
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/admin/user/update"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &body)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "update user alice email new@x.com full_name Alice")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if body["account_name"] != "alice" || body["email"] != "new@x.com" || body["full_name"] != "Alice" {
		t.Errorf("body: %+v", body)
	}
	if !strings.Contains(buf.String(), "OK: user updated") {
		t.Errorf("ack: %s", buf.String())
	}
}

func TestDispatch_UpdateGroup(t *testing.T) {
	var body map[string]any
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/group/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []GroupInfo{{ID: 7, Name: "dev"}})
		},
		{http.MethodPost, "/aa/group/update"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &body)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, `update group dev description "devs only"`)); err != nil {
		t.Fatalf("err: %v", err)
	}
	if body["id"] != float64(7) || body["description"] != "devs only" {
		t.Errorf("body: %+v", body)
	}
	if !strings.Contains(buf.String(), "OK: group updated") {
		t.Errorf("ack: %s", buf.String())
	}

	// Unresolvable group name.
	if err := d.Run(mustParse(t, "update group missing description x")); err == nil {
		t.Errorf("expected resolve error")
	}
}

// --- delete user / delete group ---

func TestDispatch_DeleteUser(t *testing.T) {
	var body map[string]any
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/authenticate/user/delete"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &body)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "delete user alice")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if body["account_name"] != "alice" {
		t.Errorf("body: %+v", body)
	}
	if !strings.Contains(buf.String(), "OK: user deleted") {
		t.Errorf("ack: %s", buf.String())
	}
}

func TestDispatch_DeleteGroup(t *testing.T) {
	var body map[string]any
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/group/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []GroupInfo{{ID: 5, Name: "dev"}})
		},
		{http.MethodPost, "/aa/group/delete"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &body)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "delete group dev")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if body["id"] != float64(5) {
		t.Errorf("body: %+v", body)
	}
	if !strings.Contains(buf.String(), "OK: group deleted") {
		t.Errorf("ack: %s", buf.String())
	}
}

// --- map/unmap remaining shapes ---

func TestDispatch_MapUserGroup(t *testing.T) {
	var body map[string]any
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/group/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []GroupInfo{{ID: 5, Name: "dev"}})
		},
		{http.MethodPost, "/aa/group/user/assign"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &body)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "map user alice group dev")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if body["username"] != "alice" || body["group_id"] != float64(5) {
		t.Errorf("body: %+v", body)
	}
}

func TestDispatch_UnmapUserGroup(t *testing.T) {
	var body map[string]any
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/group/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []GroupInfo{{ID: 5, Name: "dev"}})
		},
		{http.MethodPost, "/aa/group/user/unassign"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &body)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "unmap user alice group dev")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if body["username"] != "alice" || body["group_id"] != float64(5) {
		t.Errorf("body: %+v", body)
	}
}

func TestDispatch_UnmapUserNe(t *testing.T) {
	var body map[string]string
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/admin/ne/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []NeInfo{{ID: 10, NeName: "HTSMF01"}})
		},
		{http.MethodPost, "/aa/authorize/ne/delete"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &body)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "unmap user alice ne HTSMF01")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if body["username"] != "alice" || body["neid"] != "10" {
		t.Errorf("body: %+v", body)
	}
}

func TestDispatch_MapGroupNe(t *testing.T) {
	var body map[string]int64
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/group/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []GroupInfo{{ID: 5, Name: "dev"}})
		},
		{http.MethodGet, "/aa/admin/ne/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []NeInfo{{ID: 10, NeName: "HTSMF01"}})
		},
		{http.MethodPost, "/aa/group/ne/assign"}: func(w http.ResponseWriter, r *http.Request) {
			readJSON(t, r, &body)
			writeJSON(w, 200, map[string]string{"status": "success"})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "map group dev ne HTSMF01")); err != nil {
		t.Fatalf("err: %v", err)
	}
	if body["group_id"] != 5 || body["ne_id"] != 10 {
		t.Errorf("body: %+v", body)
	}
}

// --- error / unsupported ---

func TestDispatch_ShowUnknownEntity(t *testing.T) {
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{})
	defer done()
	cmd := &Command{Verb: "show", Entity: "widget"}
	if err := d.Run(cmd); err == nil || !strings.Contains(err.Error(), "show widget not supported") {
		t.Errorf("expected unsupported-entity error, got %v", err)
	}
}

func TestDispatch_MapUnsupportedShape(t *testing.T) {
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{})
	defer done()
	// ne can't be the subject of a map
	cmd := &Command{Verb: "map", Entity: "ne", Target: "x", Relation: "user", Related: "alice"}
	if err := d.Run(cmd); err == nil || !strings.Contains(err.Error(), "unsupported mapping") {
		t.Errorf("expected unsupported mapping, got %v", err)
	}
}

func TestDispatch_UnknownVerb(t *testing.T) {
	d, _, done := newDispatcher(t, map[route]http.HandlerFunc{})
	defer done()
	cmd := &Command{Verb: "bogus"}
	if err := d.Run(cmd); err == nil || !strings.Contains(err.Error(), "unhandled verb") {
		t.Errorf("expected unhandled-verb error, got %v", err)
	}
}

func TestDispatch_HelpUnknownTopic(t *testing.T) {
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{})
	defer done()
	_ = d.Run(mustParse(t, "help bogus"))
	if !strings.Contains(buf.String(), "no help topic") {
		t.Errorf("expected unknown-topic message, got %q", buf.String())
	}
}

func TestDispatch_GroupShow(t *testing.T) {
	d, buf, done := newDispatcher(t, map[route]http.HandlerFunc{
		{http.MethodGet, "/aa/group/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []GroupInfo{{ID: 5, Name: "dev"}})
		},
		{http.MethodPost, "/aa/group/show"}: func(w http.ResponseWriter, r *http.Request) {
			var b map[string]int64
			_ = json.NewDecoder(r.Body).Decode(&b)
			if b["id"] != 5 {
				t.Errorf("show id: %+v", b)
			}
			writeJSON(w, 200, GroupDetail{ID: 5, Name: "dev", Users: []string{"alice", "bob"}, NeIDs: []int64{10}})
		},
	})
	defer done()

	if err := d.Run(mustParse(t, "show group dev")); err != nil {
		t.Fatalf("err: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "alice, bob") {
		t.Errorf("users: %s", out)
	}
}
