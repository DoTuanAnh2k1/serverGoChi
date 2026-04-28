// Package sshcli is the v2 SSH bastion (cli-gate). It accepts operator SSH
// connections, delegates authentication to cli-mgt-svc via
// POST /aa/authenticate, then provides an interactive REPL for NE navigation
// and SSH proxying.
package sshcli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is populated from environment variables by LoadConfig.
type Config struct {
	// ListenAddr is the address the SSH server binds to (e.g. ":2223").
	ListenAddr string
	// HostKeyPath is the path to the PEM-encoded host private key. If the
	// file does not exist and AutoGenHostKey is true, an ECDSA key is
	// generated and saved there on first run.
	HostKeyPath    string
	AutoGenHostKey bool
	// MgtSvcBase is the base URL of cli-mgt-svc (e.g. "http://mgt:3000").
	MgtSvcBase string
	// HTTPTimeout controls per-request timeout when calling cli-mgt-svc.
	HTTPTimeout time.Duration
	// IdleTimeout closes a session that sends no activity for this duration
	// (0 = disabled).
	IdleTimeout time.Duration
	LogLevel    string
}

// LoadConfig reads all config from environment variables.
func LoadConfig() (Config, error) {
	c := Config{
		ListenAddr:     env("SSH_CLI_LISTEN_ADDR", ":2223"),
		HostKeyPath:    env("SSH_CLI_HOST_KEY_PATH", "/data/ssh-gate/host_key"),
		AutoGenHostKey: envBool("SSH_CLI_AUTO_GEN_HOST_KEY", true),
		MgtSvcBase:     strings.TrimRight(env("MGT_SVC_BASE", "http://localhost:3000"), "/"),
		HTTPTimeout:    envDur("SSH_CLI_HTTP_TIMEOUT", 15*time.Second),
		IdleTimeout:    envDur("SSH_CLI_IDLE_TIMEOUT", 0),
		LogLevel:       env("LOG_LEVEL", "info"),
	}
	if c.MgtSvcBase == "" {
		return c, fmt.Errorf("MGT_SVC_BASE is required")
	}
	return c, nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return def
	}
	return b
}

func envDur(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
