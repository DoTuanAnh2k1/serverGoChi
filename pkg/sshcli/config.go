package sshcli

import (
	"fmt"
	"os"
	"strings"
)

// Config is the env-driven configuration for the SSH CLI server.
type Config struct {
	ListenAddr      string
	HostKeyPath     string
	MgtSvcBase      string
	NeConfigAddr    string
	NeCommandAddr   string
	LogLevel        string
}

// LoadConfig reads environment variables and returns a populated Config or an
// error naming the first missing required variable.
func LoadConfig() (*Config, error) {
	c := &Config{
		ListenAddr:    envOr("SSH_CLI_LISTEN_ADDR", ":2223"),
		HostKeyPath:   envOr("SSH_CLI_HOST_KEY_PATH", "/data/ssh_cli_host_key"),
		MgtSvcBase:    strings.TrimRight(os.Getenv("MGT_SVC_BASE"), "/"),
		NeConfigAddr:  os.Getenv("NE_CONFIG_SSH_ADDR"),
		NeCommandAddr: os.Getenv("NE_COMMAND_SSH_ADDR"),
		LogLevel:      envOr("LOG_LEVEL", "info"),
	}
	if c.MgtSvcBase == "" {
		return nil, fmt.Errorf("MGT_SVC_BASE is required")
	}
	// ne-config and ne-command addrs are optional — the corresponding menu
	// item will report "not configured" if missing.
	return c, nil
}

func envOr(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}
