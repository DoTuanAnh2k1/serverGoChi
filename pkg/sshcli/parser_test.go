package sshcli

import (
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	cases := []struct {
		in   string
		want []string
		err  bool
	}{
		{"", nil, false},
		{"   ", nil, false},
		{"show user", []string{"show", "user"}, false},
		{"set user name alice password pass123", []string{"set", "user", "name", "alice", "password", "pass123"}, false},
		{`update user alice full_name "John Doe"`, []string{"update", "user", "alice", "full_name", "John Doe"}, false},
		{`set ne description "HCM SMF 01"`, []string{"set", "ne", "description", "HCM SMF 01"}, false},
		{`full_name "he said \"hi\""`, []string{"full_name", `he said "hi"`}, false},
		{`full_name "oops`, nil, true},
	}
	for _, tc := range cases {
		got, err := Tokenize(tc.in)
		if tc.err {
			if err == nil {
				t.Errorf("Tokenize(%q) expected error, got %v", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("Tokenize(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("Tokenize(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestParse_SimpleVerbs(t *testing.T) {
	c, err := Parse("exit")
	if err != nil || c.Verb != "exit" {
		t.Fatalf("exit: %+v err=%v", c, err)
	}
	c, err = Parse("quit")
	if err != nil || c.Verb != "quit" {
		t.Fatalf("quit: %+v err=%v", c, err)
	}
	c, err = Parse("help")
	if err != nil || c.Verb != "help" || c.Target != "" {
		t.Fatalf("help: %+v err=%v", c, err)
	}
	c, err = Parse("help set")
	if err != nil || c.Target != "set" {
		t.Fatalf("help set: %+v err=%v", c, err)
	}
	if _, err := Parse("exit now"); err == nil {
		t.Errorf("exit with args should error")
	}
}

func TestParse_Show(t *testing.T) {
	c, err := Parse("show user")
	if err != nil || c.Verb != "show" || c.Entity != "user" || c.Target != "" {
		t.Fatalf("show user: %+v err=%v", c, err)
	}
	c, err = Parse("show user alice")
	if err != nil || c.Target != "alice" {
		t.Fatalf("show user alice: %+v err=%v", c, err)
	}
	if _, err := Parse("show user alice bob"); err == nil {
		t.Errorf("show with too many args should error")
	}
	if _, err := Parse("show"); err == nil {
		t.Errorf("show without entity should error")
	}
	if _, err := Parse("show widget"); err == nil {
		t.Errorf("show with unknown entity should error")
	}
}

func TestParse_SetPairs(t *testing.T) {
	c, err := Parse("set user name alice password pass123 email a@b.c")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := map[string]string{"name": "alice", "password": "pass123", "email": "a@b.c"}
	if !reflect.DeepEqual(c.Fields, want) {
		t.Errorf("fields: got %v want %v", c.Fields, want)
	}
	wantOrder := []string{"name", "password", "email"}
	if !reflect.DeepEqual(c.FieldOrder, wantOrder) {
		t.Errorf("order: got %v want %v", c.FieldOrder, wantOrder)
	}
	if _, err := Parse("set user name"); err == nil {
		t.Errorf("odd number of tokens should error")
	}
	if _, err := Parse("set user name a name b"); err == nil {
		t.Errorf("duplicate field should error")
	}
	if _, err := Parse("set user"); err == nil {
		t.Errorf("empty pairs should error")
	}
}

func TestParse_Update(t *testing.T) {
	c, err := Parse("update user alice email new@b.c")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if c.Target != "alice" || c.Fields["email"] != "new@b.c" {
		t.Errorf("got %+v", c)
	}
	if _, err := Parse("update user alice email"); err == nil {
		t.Errorf("update with no value should error")
	}
	if _, err := Parse("update user alice"); err == nil {
		t.Errorf("update with no pairs should error")
	}
}

func TestParse_Delete(t *testing.T) {
	c, err := Parse("delete user alice")
	if err != nil || c.Target != "alice" {
		t.Fatalf("delete user alice: %+v err=%v", c, err)
	}
	if _, err := Parse("delete user"); err == nil {
		t.Errorf("delete without target should error")
	}
	if _, err := Parse("delete user a b"); err == nil {
		t.Errorf("delete with extra args should error")
	}
}

func TestParse_MapUnmap(t *testing.T) {
	c, err := Parse("map user alice ne HTSMF01")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if c.Entity != "user" || c.Target != "alice" || c.Relation != "ne" || c.Related != "HTSMF01" {
		t.Errorf("got %+v", c)
	}
	c, err = Parse("map user alice group dev")
	if err != nil || c.Relation != "group" || c.Related != "dev" {
		t.Fatalf("map user alice group dev: %+v err=%v", c, err)
	}
	c, err = Parse("unmap group dev ne HTSMF01")
	if err != nil || c.Verb != "unmap" || c.Entity != "group" || c.Relation != "ne" {
		t.Fatalf("unmap group ne: %+v err=%v", c, err)
	}
	if _, err := Parse("map user alice bogus HTSMF01"); err == nil {
		t.Errorf("bad relation should error")
	}
	if _, err := Parse("map group dev group other"); err == nil {
		t.Errorf("group->group map should error")
	}
	if _, err := Parse("map ne foo ne bar"); err == nil {
		t.Errorf("ne as subject should error")
	}
	if _, err := Parse("map user alice ne"); err == nil {
		t.Errorf("missing related should error")
	}
}

func TestParse_UnknownVerb(t *testing.T) {
	if _, err := Parse("frobnicate user alice"); err == nil {
		t.Errorf("unknown verb should error")
	}
}

func TestParse_Empty(t *testing.T) {
	c, err := Parse("")
	if err != nil || c != nil {
		t.Errorf("empty line: got %+v err=%v, want nil/nil", c, err)
	}
	c, err = Parse("    ")
	if err != nil || c != nil {
		t.Errorf("whitespace: got %+v err=%v, want nil/nil", c, err)
	}
}

func TestParse_CaseInsensitiveVerb(t *testing.T) {
	c, err := Parse("SHOW User alice")
	if err != nil || c.Verb != "show" || c.Entity != "user" || c.Target != "alice" {
		t.Errorf("case: %+v err=%v", c, err)
	}
}
