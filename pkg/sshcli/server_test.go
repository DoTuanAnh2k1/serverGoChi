package sshcli

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// TestServer_EndToEnd spins up a fake mgt-svc + our SSH server and drives it
// with a real ssh.Client. We authenticate, land in the menu, switch to
// cli-config, issue "show user", and exit.
func TestServer_EndToEnd(t *testing.T) {
	adminTok := fakeJWT("admin")
	mgt := newFakeMgt(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/authenticate"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, map[string]any{
				"status":        "success",
				"response_data": adminTok,
				"response_code": "200",
			})
		},
		{http.MethodGet, "/aa/admin/user/list"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, []UserInfo{
				{AccountID: 1, AccountName: "alice", AccountType: 1, IsEnable: true},
				{AccountID: 2, AccountName: "bob", AccountType: 2, IsEnable: false},
			})
		},
	})
	defer mgt.Close()

	dir := t.TempDir()
	cfg := &Config{
		ListenAddr:  "127.0.0.1:0",
		HostKeyPath: filepath.Join(dir, "hk"),
		MgtSvcBase:  mgt.URL,
	}

	s, err := NewServer(cfg)
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	lst, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = serveOnListener(s, lst, ctx) }()

	clientCfg := &ssh.ClientConfig{
		User:            "alice",
		Auth:            []ssh.AuthMethod{ssh.Password("pw")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         3 * time.Second,
	}
	conn, err := ssh.Dial("tcp", lst.Addr().String(), clientCfg)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	sess, err := conn.NewSession()
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	defer sess.Close()

	if err := sess.RequestPty("xterm", 40, 120, ssh.TerminalModes{}); err != nil {
		t.Fatalf("pty: %v", err)
	}
	stdin, _ := sess.StdinPipe()
	stdout, _ := sess.StdoutPipe()
	if err := sess.Shell(); err != nil {
		t.Fatalf("shell: %v", err)
	}

	// Collect output in the background.
	outC := make(chan string, 1)
	go func() {
		all := &strings.Builder{}
		buf := bufio.NewReader(stdout)
		for {
			b, err := buf.ReadByte()
			if err != nil {
				break
			}
			all.WriteByte(b)
		}
		outC <- all.String()
	}()

	// Give the server a moment to send the banner.
	time.Sleep(200 * time.Millisecond)

	commands := []string{
		"cli-config\r",
		"show user\r",
		"exit\r",
		"exit\r",
	}
	for _, c := range commands {
		_, _ = io.WriteString(stdin, c)
		time.Sleep(200 * time.Millisecond)
	}
	_ = stdin.Close()
	_ = sess.Wait()

	var out string
	select {
	case out = <-outC:
	case <-time.After(2 * time.Second):
		t.Fatal("output timeout")
	}

	for _, want := range []string{"Welcome alice", "cli-config mode", "alice", "bob", "bye."} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q; got:\n%s", want, out)
		}
	}
}

// serveOnListener is a tiny variant of Server.ListenAndServe that takes an
// existing listener — handy for tests that want to pick a random port.
func serveOnListener(s *Server, lst net.Listener, ctx context.Context) error {
	go func() {
		<-ctx.Done()
		_ = lst.Close()
	}()
	for {
		conn, err := lst.Accept()
		if err != nil {
			return err
		}
		go s.handleConn(conn)
	}
}

func TestServer_AuthFailure(t *testing.T) {
	mgt := newFakeMgt(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/authenticate"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 401, map[string]string{"status": "error"})
		},
	})
	defer mgt.Close()

	dir := t.TempDir()
	cfg := &Config{
		ListenAddr:  "127.0.0.1:0",
		HostKeyPath: filepath.Join(dir, "hk"),
		MgtSvcBase:  mgt.URL,
	}
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}

	lst, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = serveOnListener(s, lst, ctx) }()

	clientCfg := &ssh.ClientConfig{
		User:            "alice",
		Auth:            []ssh.AuthMethod{ssh.Password("bad")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         2 * time.Second,
	}
	if _, err := ssh.Dial("tcp", lst.Addr().String(), clientCfg); err == nil {
		t.Errorf("expected auth failure on dial")
	}
}

// Normal users (role=="user") are now admitted — the mode menu filters
// cli-config out for them. This test verifies the handshake succeeds; the
// mode-filter behavior is covered by TestAvailableModes.
func TestServer_AcceptNormalUser(t *testing.T) {
	userTok := fakeJWT("user")
	mgt := newFakeMgt(t, map[route]http.HandlerFunc{
		{http.MethodPost, "/aa/authenticate"}: func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, 200, map[string]any{
				"status": "success", "response_data": userTok, "response_code": "200",
			})
		},
	})
	defer mgt.Close()

	dir := t.TempDir()
	cfg := &Config{
		ListenAddr:  "127.0.0.1:0",
		HostKeyPath: filepath.Join(dir, "hk"),
		MgtSvcBase:  mgt.URL,
	}
	s, err := NewServer(cfg)
	if err != nil {
		t.Fatal(err)
	}

	lst, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { _ = serveOnListener(s, lst, ctx) }()

	clientCfg := &ssh.ClientConfig{
		User:            "alice",
		Auth:            []ssh.AuthMethod{ssh.Password("pw")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         2 * time.Second,
	}
	conn, err := ssh.Dial("tcp", lst.Addr().String(), clientCfg)
	if err != nil {
		t.Fatalf("normal user should be accepted: %v", err)
	}
	_ = conn.Close()
}

func TestAvailableModes(t *testing.T) {
	admin := availableModes("admin")
	if len(admin) != 3 {
		t.Errorf("admin should see 3 modes, got %v", admin)
	}
	user := availableModes("user")
	if len(user) != 2 {
		t.Errorf("normal user should see 2 modes, got %v", user)
	}
	for _, m := range user {
		if m == "cli-config" {
			t.Errorf("normal user must not see cli-config, got %v", user)
		}
	}
}
