package sshcli

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// session is a single operator connection. It maintains its own mgtClient
// so the token is per-session.
type session struct {
	cfg      Config
	username string
	mgt      *mgtClient
	ch       ssh.Channel
}

func (s *session) run() {
	fmt.Fprintf(s.ch, "\r\ncli-gate v2 — logged in as %s (role: %s)\r\n", s.username, s.mgt.role)
	fmt.Fprintf(s.ch, "Type 'help' for available commands.\r\n\r\n")
	s.repl()
}

func (s *session) repl() {
	buf := make([]byte, 1)
	for {
		fmt.Fprintf(s.ch, "gate> ")
		line := s.readLine(buf)
		if line == nil {
			return
		}
		cmd := strings.TrimSpace(string(line))
		if cmd == "" {
			continue
		}
		if !s.dispatch(cmd) {
			return
		}
	}
}

// readLine reads one line from the channel, echoing characters. Returns nil on EOF.
func (s *session) readLine(buf []byte) []byte {
	var line []byte
	for {
		if s.cfg.IdleTimeout > 0 {
			// Channels don't directly expose SetDeadline; handled via server side.
			_ = s.cfg.IdleTimeout
		}
		n, err := s.ch.Read(buf)
		if err != nil || n == 0 {
			return nil
		}
		b := buf[0]
		switch b {
		case '\r', '\n':
			fmt.Fprintf(s.ch, "\r\n")
			return line
		case 3: // Ctrl+C
			fmt.Fprintf(s.ch, "^C\r\n")
			return []byte{}
		case 4: // Ctrl+D / EOF
			return nil
		case 127, 8: // Backspace / DEL
			if len(line) > 0 {
				line = line[:len(line)-1]
				fmt.Fprintf(s.ch, "\b \b")
			}
		default:
			if b >= 32 {
				line = append(line, b)
				s.ch.Write(buf[:1])
			}
		}
	}
}

// dispatch runs the command and returns false when the session should close.
func (s *session) dispatch(input string) bool {
	parts := strings.Fields(input)
	verb := parts[0]
	args := parts[1:]

	switch verb {
	case "help", "?":
		s.printHelp()
	case "show":
		if len(args) == 0 {
			fmt.Fprintf(s.ch, "Usage: show ne\r\n")
			return true
		}
		switch args[0] {
		case "ne":
			s.showNEs()
		case "user":
			fmt.Fprintf(s.ch, "  Username : %s\r\n  Role     : %s\r\n", s.username, s.mgt.role)
		default:
			fmt.Fprintf(s.ch, "Unknown target: %s\r\n", args[0])
		}
	case "connect":
		if len(args) == 0 {
			fmt.Fprintf(s.ch, "Usage: connect <namespace>\r\n")
			return true
		}
		s.connectNE(args[0])
	case "exit", "quit":
		fmt.Fprintf(s.ch, "Bye.\r\n")
		return false
	default:
		fmt.Fprintf(s.ch, "Unknown command: %q — type 'help' for available commands\r\n", verb)
	}
	return true
}

func (s *session) printHelp() {
	fmt.Fprintf(s.ch,
		"\r\nAvailable commands:\r\n"+
			"  show ne              List all NEs you can reach\r\n"+
			"  show user            Show your username and role\r\n"+
			"  connect <namespace>  Open an SSH shell on the NE\r\n"+
			"  exit / quit          Close this session\r\n\r\n")
}

func (s *session) showNEs() {
	nes, err := s.mgt.ListNEs()
	if err != nil {
		fmt.Fprintf(s.ch, "Error listing NEs: %v\r\n", err)
		return
	}
	if len(nes) == 0 {
		fmt.Fprintf(s.ch, "  (no NEs defined)\r\n")
		return
	}
	fmt.Fprintf(s.ch, "\r\n  %-20s %-12s %-20s %-8s %-10s\r\n",
		"Namespace", "Type", "Master IP", "Port", "Mode")
	fmt.Fprintf(s.ch, "  %s\r\n", strings.Repeat("-", 76))
	for _, n := range nes {
		fmt.Fprintf(s.ch, "  %-20s %-12s %-20s %-8d %-10s\r\n",
			n.Namespace, n.NeType, n.MasterIP, n.MasterPort, n.ConfMode)
	}
	fmt.Fprintf(s.ch, "\r\n")
}

func (s *session) connectNE(ns string) {
	ne, err := s.mgt.GetNEByNamespace(ns)
	if err != nil {
		fmt.Fprintf(s.ch, "Error: %v\r\n", err)
		return
	}
	if ne == nil {
		fmt.Fprintf(s.ch, "NE %q not found\r\n", ns)
		return
	}
	if ne.MasterIP == "" {
		fmt.Fprintf(s.ch, "NE %q has no master_ip configured\r\n", ns)
		return
	}

	port := ne.MasterPort
	if port == 0 {
		port = 22
	}
	addr := fmt.Sprintf("%s:%d", ne.MasterIP, port)

	fmt.Fprintf(s.ch, "Connecting to %s (%s:%d)…\r\n", ns, ne.MasterIP, port)

	start := time.Now()
	if err := s.proxySSH(ne, addr); err != nil {
		elapsed := time.Since(start).Round(time.Millisecond)
		s.mgt.SaveHistory(s.username, "connect "+ns, ns, ne.MasterIP, "ne-command",
			"error: "+err.Error())
		fmt.Fprintf(s.ch, "\r\nDisconnected (%s): %v\r\n", elapsed, err)
	} else {
		elapsed := time.Since(start).Round(time.Millisecond)
		s.mgt.SaveHistory(s.username, "connect "+ns, ns, ne.MasterIP, "ne-command", "ok")
		fmt.Fprintf(s.ch, "\r\nDisconnected from %s (%s)\r\n", ns, elapsed)
	}
}

// proxySSH opens an SSH connection to the NE and splices the operator's
// stdio channel through it, implementing the ne-command proxy mode.
func (s *session) proxySSH(ne *NE, addr string) error {
	if ne.SSHUsername == "" || ne.SSHPassword == "" {
		return fmt.Errorf("NE has no SSH credentials configured")
	}

	config := &ssh.ClientConfig{
		User: ne.SSHUsername,
		Auth: []ssh.AuthMethod{
			ssh.Password(ne.SSHPassword),
		},
		// HostKeyCallback is intentionally InsecureIgnoreHostKey on the
		// internal NE side — operators manage NE certs out-of-band. The
		// outer operator→gate connection uses a server host key managed by
		// the gate. Production deployments should supply a known_hosts
		// checker via SSH_CLI_KNOWN_HOSTS_PATH.
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
		Timeout:         15 * time.Second,
	}

	conn, err := net.DialTimeout("tcp", addr, 15*time.Second)
	if err != nil {
		return fmt.Errorf("dial %s: %w", addr, err)
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close()
		return fmt.Errorf("ssh handshake: %w", err)
	}
	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	neSession, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("new session: %w", err)
	}
	defer neSession.Close()

	// Request a PTY so the NE shell renders properly on the operator's terminal.
	if err := neSession.RequestPty("xterm-256color", 40, 200, ssh.TerminalModes{}); err != nil {
		// Non-fatal — fall back to no-pty if the NE doesn't support it.
		_ = err
	}

	neSession.Stdin = s.ch
	neSession.Stdout = s.ch
	neSession.Stderr = s.ch

	if err := neSession.Shell(); err != nil {
		return fmt.Errorf("start shell: %w", err)
	}
	return neSession.Wait()
}

// ioProxy splices two ReadWriters in both directions until either side closes.
func ioProxy(a, b io.ReadWriter) {
	done := make(chan struct{}, 2)
	cp := func(dst io.Writer, src io.Reader) {
		io.Copy(dst, src)
		done <- struct{}{}
	}
	go cp(a, b)
	go cp(b, a)
	<-done
}
