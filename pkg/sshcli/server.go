package sshcli

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// Server is the top-level SSH listener that authenticates users against
// mgt-svc and dispatches the session to the mode menu.
type Server struct {
	Cfg       *Config
	sshConfig *ssh.ServerConfig
}

// NewServer builds a Server with host key loaded or generated at HostKeyPath.
func NewServer(cfg *Config) (*Server, error) {
	signer, err := loadOrCreateHostKey(cfg.HostKeyPath)
	if err != nil {
		return nil, fmt.Errorf("host key: %w", err)
	}
	s := &Server{Cfg: cfg}
	s.sshConfig = &ssh.ServerConfig{
		PasswordCallback: s.passwordAuth,
	}
	s.sshConfig.AddHostKey(signer)
	return s, nil
}

// ListenAndServe starts accepting SSH connections. It blocks until ctx is done.
func (s *Server) ListenAndServe(ctx context.Context) error {
	lst, err := net.Listen("tcp", s.Cfg.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.Cfg.ListenAddr, err)
	}
	log.Infof("ssh-cli server listening on %s", s.Cfg.ListenAddr)
	go func() {
		<-ctx.Done()
		_ = lst.Close()
	}()

	var wg sync.WaitGroup
	for {
		conn, err := lst.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			log.Warnf("accept: %v", err)
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.handleConn(conn)
		}()
	}
	wg.Wait()
	return nil
}

// passwordAuth authenticates using mgt-svc and stashes the bound MgtClient in
// the connection metadata so the session goroutine can reuse it.
func (s *Server) passwordAuth(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	username := c.User()
	password := string(pass)
	client := NewMgtClient(s.Cfg.MgtSvcBase)
	if err := client.Authenticate(username, password); err != nil {
		log.Warnf("auth failed for %s@%s: %v", username, c.RemoteAddr(), err)
		return nil, fmt.Errorf("authentication failed")
	}
	log.Infof("auth ok: %s@%s (role=%s)", username, c.RemoteAddr(), client.Role)

	perms := &ssh.Permissions{
		Extensions: map[string]string{
			"username": username,
			"password": password,
			"role":     client.Role,
			"token":    client.Token,
		},
	}
	// Stash the prebuilt client in an in-memory map keyed by session-id so
	// we don't re-auth on session setup. The Extensions map survives to the
	// session handler via conn.Permissions.
	clientMu.Lock()
	clientByToken[client.Token] = client
	clientMu.Unlock()
	return perms, nil
}

var (
	clientMu      sync.Mutex
	clientByToken = map[string]*MgtClient{}
)

func clientForToken(tok string) *MgtClient {
	clientMu.Lock()
	defer clientMu.Unlock()
	c := clientByToken[tok]
	delete(clientByToken, tok)
	return c
}

func (s *Server) handleConn(nconn net.Conn) {
	defer nconn.Close()
	conn, chans, reqs, err := ssh.NewServerConn(nconn, s.sshConfig)
	if err != nil {
		log.Debugf("handshake failed from %s: %v", nconn.RemoteAddr(), err)
		return
	}
	defer conn.Close()
	go ssh.DiscardRequests(reqs)

	ext := conn.Permissions.Extensions
	username := ext["username"]
	password := ext["password"]
	token := ext["token"]
	client := clientForToken(token)
	if client == nil {
		// Stashed client was evicted between PasswordCallback and channel
		// setup; rebuild from the token without a re-auth round-trip.
		client = NewMgtClient(s.Cfg.MgtSvcBase)
		client.Token = token
		client.Role = conn.Permissions.Extensions["role"]
	}

	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			_ = newChan.Reject(ssh.UnknownChannelType, "only session channels supported")
			continue
		}
		ch, chReqs, err := newChan.Accept()
		if err != nil {
			log.Warnf("accept channel: %v", err)
			continue
		}
		go s.handleSession(ch, chReqs, client, username, password)
	}
}

func (s *Server) handleSession(ch ssh.Channel, reqs <-chan *ssh.Request, client *MgtClient, username, password string) {
	defer ch.Close()

	var ptyTerm string
	var ptyW, ptyH uint32
	var ptyModes map[uint8]uint32
	winchgs := make(chan WindowSize, 8)
	shellStarted := false

	go func() {
		for req := range reqs {
			switch req.Type {
			case "pty-req":
				term, w, h, modes, ok := parsePtyReq(req.Payload)
				if ok {
					ptyTerm, ptyW, ptyH, ptyModes = term, w, h, modes
				}
				_ = req.Reply(ok, nil)
			case "window-change":
				w, h, ok := parseWindowChange(req.Payload)
				if ok {
					select {
					case winchgs <- WindowSize{Width: w, Height: h}:
					default:
					}
				}
				_ = req.Reply(ok, nil)
			case "shell":
				if !shellStarted {
					shellStarted = true
					_ = req.Reply(true, nil)
				} else {
					_ = req.Reply(false, nil)
				}
			case "exec":
				// We don't support non-interactive exec; tell the client.
				_ = req.Reply(false, nil)
			default:
				_ = req.Reply(false, nil)
			}
		}
	}()

	// Wait for shell request by spinning briefly — the request arrives
	// before meaningful data; for simplicity we start the session immediately
	// and let the goroutine above ack the shell req when it arrives. This
	// matches common Go SSH server patterns.

	runner := &SessionRunner{
		Client:        client,
		Username:      username,
		Password:      password,
		NeConfigAddr:  s.Cfg.NeConfigAddr,
		NeCommandAddr: s.Cfg.NeCommandAddr,
		PTYTerm:       ptyTerm,
		PTYWidth:      ptyW,
		PTYHeight:     ptyH,
		PTYModes:      ptyModes,
		WindowChanges: winchgs,
	}
	if err := runner.Run(ch); err != nil && !errors.Is(err, io.EOF) {
		log.Warnf("session for %s ended: %v", username, err)
	}
	close(winchgs)
}

// parsePtyReq parses the RFC 4254 payload: string term, uint32 w, uint32 h,
// uint32 wPx, uint32 hPx, string modes.
func parsePtyReq(payload []byte) (string, uint32, uint32, map[uint8]uint32, bool) {
	term, payload, ok := parseSSHString(payload)
	if !ok {
		return "", 0, 0, nil, false
	}
	if len(payload) < 16 {
		return "", 0, 0, nil, false
	}
	w := binary.BigEndian.Uint32(payload[0:4])
	h := binary.BigEndian.Uint32(payload[4:8])
	payload = payload[16:]
	modesRaw, _, ok := parseSSHString(payload)
	if !ok {
		return term, w, h, nil, true
	}
	modes := decodeTerminalModes([]byte(modesRaw))
	return term, w, h, modes, true
}

func parseWindowChange(payload []byte) (uint32, uint32, bool) {
	if len(payload) < 16 {
		return 0, 0, false
	}
	w := binary.BigEndian.Uint32(payload[0:4])
	h := binary.BigEndian.Uint32(payload[4:8])
	return w, h, true
}

func parseSSHString(b []byte) (string, []byte, bool) {
	if len(b) < 4 {
		return "", b, false
	}
	n := binary.BigEndian.Uint32(b[0:4])
	if len(b) < int(4+n) {
		return "", b, false
	}
	return string(b[4 : 4+n]), b[4+n:], true
}

// decodeTerminalModes parses the encoded modes blob from a pty-req payload.
// Format: sequence of (opcode uint8, argument uint32) ending at opcode 0 (TTY_OP_END).
func decodeTerminalModes(b []byte) map[uint8]uint32 {
	out := map[uint8]uint32{}
	for len(b) >= 5 {
		op := b[0]
		if op == 0 {
			return out
		}
		val := binary.BigEndian.Uint32(b[1:5])
		out[op] = val
		b = b[5:]
	}
	return out
}

func loadOrCreateHostKey(path string) (ssh.Signer, error) {
	if raw, err := os.ReadFile(path); err == nil {
		return ssh.ParsePrivateKey(raw)
	}
	// Generate a new ed25519 key pair and persist it.
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	pemBlock, err := ssh.MarshalPrivateKey(priv, "ssh-cli host key")
	if err != nil {
		return nil, err
	}
	out := pem.EncodeToMemory(pemBlock)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(out)
}
