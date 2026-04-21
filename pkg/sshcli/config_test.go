package sshcli

import (
	"strings"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	for _, k := range []string{"SSH_CLI_LISTEN_ADDR", "SSH_CLI_HOST_KEY_PATH", "NE_CONFIG_SSH_ADDR", "NE_COMMAND_SSH_ADDR", "LOG_LEVEL"} {
		t.Setenv(k, "")
	}
	t.Setenv("MGT_SVC_BASE", "http://mgt:3000/")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cfg.ListenAddr != ":2223" {
		t.Errorf("ListenAddr default: %q", cfg.ListenAddr)
	}
	if cfg.HostKeyPath != "/data/ssh_cli_host_key" {
		t.Errorf("HostKeyPath default: %q", cfg.HostKeyPath)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel default: %q", cfg.LogLevel)
	}
	// Trailing slash is trimmed.
	if cfg.MgtSvcBase != "http://mgt:3000" {
		t.Errorf("MgtSvcBase: %q", cfg.MgtSvcBase)
	}
	if cfg.NeConfigAddr != "" || cfg.NeCommandAddr != "" {
		t.Errorf("optional addrs should be empty: %+v", cfg)
	}
}

func TestLoadConfig_AllSet(t *testing.T) {
	t.Setenv("SSH_CLI_LISTEN_ADDR", ":9999")
	t.Setenv("SSH_CLI_HOST_KEY_PATH", "/tmp/hk")
	t.Setenv("MGT_SVC_BASE", "http://mgt:3000")
	t.Setenv("NE_CONFIG_SSH_ADDR", "ne-config:22")
	t.Setenv("NE_COMMAND_SSH_ADDR", "ne-command:22")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cfg.ListenAddr != ":9999" || cfg.HostKeyPath != "/tmp/hk" || cfg.LogLevel != "debug" {
		t.Errorf("cfg: %+v", cfg)
	}
	if cfg.NeConfigAddr != "ne-config:22" || cfg.NeCommandAddr != "ne-command:22" {
		t.Errorf("addrs: %+v", cfg)
	}
}

func TestLoadConfig_MissingMgtBase(t *testing.T) {
	t.Setenv("MGT_SVC_BASE", "")
	_, err := LoadConfig()
	if err == nil || !strings.Contains(err.Error(), "MGT_SVC_BASE") {
		t.Errorf("expected MGT_SVC_BASE error, got %v", err)
	}
}
