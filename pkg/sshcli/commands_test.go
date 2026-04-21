package sshcli

import (
	"testing"
)

func TestNormalizedFields_UserSetHappy(t *testing.T) {
	raw := map[string]string{"name": "alice", "password": "pass123", "email": "a@b.c"}
	order := []string{"name", "password", "email"}
	got, canonical, err := NormalizedFields("user", raw, order, true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got["account_name"] != "alice" || got["password"] != "pass123" || got["email"] != "a@b.c" {
		t.Errorf("got %+v", got)
	}
	if canonical[0] != "account_name" {
		t.Errorf("canonical order: %v", canonical)
	}
}

func TestNormalizedFields_UserMissingRequired(t *testing.T) {
	raw := map[string]string{"name": "alice"}
	order := []string{"name"}
	if _, _, err := NormalizedFields("user", raw, order, true); err == nil {
		t.Errorf("expected error for missing password")
	}
}

func TestNormalizedFields_UpdatePartialOK(t *testing.T) {
	raw := map[string]string{"email": "new@b.c"}
	order := []string{"email"}
	got, _, err := NormalizedFields("user", raw, order, false)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got["email"] != "new@b.c" {
		t.Errorf("got %+v", got)
	}
}

func TestNormalizedFields_AccountTypeEnum(t *testing.T) {
	raw := map[string]string{"name": "a", "password": "p", "account_type": "1"}
	order := []string{"name", "password", "account_type"}
	got, _, err := NormalizedFields("user", raw, order, true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got["account_type"] != 1 {
		t.Errorf("account_type should be int 1, got %v (%T)", got["account_type"], got["account_type"])
	}
	raw["account_type"] = "5"
	if _, _, err := NormalizedFields("user", raw, order, true); err == nil {
		t.Errorf("account_type=5 should error")
	}
}

func TestNormalizedFields_NeSetHappy(t *testing.T) {
	raw := map[string]string{
		"ne_name":              "HTSMF01",
		"namespace":            "hcm-5gc",
		"conf_master_ip":       "10.0.0.1",
		"conf_port_master_tcp": "8080",
		"command_url":          "http://10.0.0.1:8080",
	}
	order := []string{"ne_name", "namespace", "conf_master_ip", "conf_port_master_tcp", "command_url"}
	got, _, err := NormalizedFields("ne", raw, order, true)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got["conf_port_master_tcp"] != 8080 {
		t.Errorf("port should be int, got %v (%T)", got["conf_port_master_tcp"], got["conf_port_master_tcp"])
	}
}

func TestNormalizedFields_NeMissingRequired(t *testing.T) {
	raw := map[string]string{"ne_name": "x"}
	order := []string{"ne_name"}
	if _, _, err := NormalizedFields("ne", raw, order, true); err == nil {
		t.Errorf("expected error for missing ne required fields")
	}
}

func TestNormalizedFields_NeConfMode(t *testing.T) {
	raw := map[string]string{
		"ne_name":              "x",
		"namespace":            "n",
		"conf_master_ip":       "1.2.3.4",
		"conf_port_master_tcp": "1",
		"command_url":          "u",
		"conf_mode":            "BOGUS",
	}
	order := []string{"ne_name", "namespace", "conf_master_ip", "conf_port_master_tcp", "command_url", "conf_mode"}
	if _, _, err := NormalizedFields("ne", raw, order, true); err == nil {
		t.Errorf("BOGUS conf_mode should error")
	}
	raw["conf_mode"] = "SSH"
	if _, _, err := NormalizedFields("ne", raw, order, true); err != nil {
		t.Errorf("SSH conf_mode should pass: %v", err)
	}
}

func TestNormalizedFields_UnknownField(t *testing.T) {
	raw := map[string]string{"name": "x", "password": "p", "bogus": "v"}
	order := []string{"name", "password", "bogus"}
	if _, _, err := NormalizedFields("user", raw, order, true); err == nil {
		t.Errorf("unknown field should error")
	}
}

func TestNormalizedFields_IntParseFail(t *testing.T) {
	raw := map[string]string{
		"ne_name":              "x",
		"namespace":            "n",
		"conf_master_ip":       "1.2.3.4",
		"conf_port_master_tcp": "notanint",
		"command_url":          "u",
	}
	order := []string{"ne_name", "namespace", "conf_master_ip", "conf_port_master_tcp", "command_url"}
	if _, _, err := NormalizedFields("ne", raw, order, true); err == nil {
		t.Errorf("non-int port should error")
	}
}

func TestNormalizedFields_AliasConflict(t *testing.T) {
	raw := map[string]string{"name": "a", "account_name": "b", "password": "p"}
	order := []string{"name", "account_name", "password"}
	if _, _, err := NormalizedFields("user", raw, order, true); err == nil {
		t.Errorf("name + account_name alias conflict should error")
	}
}

func TestFieldNames_User(t *testing.T) {
	names := FieldNames("user")
	if len(names) == 0 {
		t.Fatal("empty")
	}
	// spot check canonical fields
	want := []string{"account_name", "password", "email", "full_name", "account_type"}
	for _, w := range want {
		if !containsString(names, w) {
			t.Errorf("expected %q in %v", w, names)
		}
	}
}

func TestFieldAliasNames(t *testing.T) {
	names := FieldAliasNames("user")
	// both alias and canonical should appear
	if !containsString(names, "name") || !containsString(names, "account_name") {
		t.Errorf("got %v", names)
	}
}
