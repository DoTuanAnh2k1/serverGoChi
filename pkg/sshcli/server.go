package sshcli

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// Server is the SSH gate. One instance handles all operator connections.
type Server struct {
	cfg       Config
	sshConfig *ssh.ServerConfig

	// tokenCache holds the mgt JWT for each active TCP connection, keyed by
	// remote address string. Populated in PasswordCallback, consumed once
	// in handleSession. Entries are deleted after the session ends.
	tokenCache sync.Map
}

// NewServer creates and configures the SSH server. It loads (or generates)
// the host key and wires up the password authentication callback.
func NewServer(cfg Config) (*Server, error) {
	hostKey, err := loadOrGenHostKey(cfg)
	if err != nil {
		return nil, fmt.Errorf("host key: %w", err)
	}

	srv := &Server{cfg: cfg}
	srv.sshConfig = &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			mgt := newMgtClient(cfg.MgtSvcBase, cfg.HTTPTimeout)
			if err := mgt.Authenticate(c.User(), string(pass)); err != nil {
				logrus.WithField("user", c.User()).Warnf("gate: auth failed: %v", err)
				return nil, fmt.Errorf("authentication failed")
			}
			logrus.WithField("user", c.User()).WithField("role", mgt.role).
				Info("gate: authenticated")
			// Cache token+role so handleSession can use them without
			// re-authenticating. Key is remote address — unique per TCP
			// connection.
			srv.tokenCache.Store(c.RemoteAddr().String(), mgt)
			return &ssh.Permissions{
				Extensions: map[string]string{"role": mgt.role},
			}, nil
		},
	}
	srv.sshConfig.AddHostKey(hostKey)
	return srv, nil
}

// ListenAndServe starts accepting connections and blocks until ctx is done.
func (s *Server) ListenAndServe(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.cfg.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.cfg.ListenAddr, err)
	}
	logrus.Infof("gate: listening on %s → mgt=%s", s.cfg.ListenAddr, s.cfg.MgtSvcBase)

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				logrus.Errorf("gate: accept: %v", err)
				continue
			}
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	remoteAddr := conn.RemoteAddr().String()

	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.sshConfig)
	if err != nil {
		s.tokenCache.Delete(remoteAddr)
		logrus.Debugf("gate: handshake %s: %v", remoteAddr, err)
		return
	}
	defer func() {
		sshConn.Close()
		s.tokenCache.Delete(remoteAddr)
	}()

	go ssh.DiscardRequests(reqs)

	for newChan := range chans {
		if newChan.ChannelType() != "session" {
			newChan.Reject(ssh.UnknownChannelType, "unsupported channel type")
			continue
		}
		ch, requests, err := newChan.Accept()
		if err != nil {
			return
		}
		// Retrieve the mgtClient that PasswordCallback stored.
		var mgt *mgtClient
		if v, ok := s.tokenCache.Load(remoteAddr); ok {
			mgt = v.(*mgtClient)
		} else {
			mgt = newMgtClient(s.cfg.MgtSvcBase, s.cfg.HTTPTimeout)
		}
		go s.handleSession(sshConn, ch, requests, mgt)
	}
}

func (s *Server) handleSession(conn *ssh.ServerConn, ch ssh.Channel, requests <-chan *ssh.Request, mgt *mgtClient) {
	defer ch.Close()

	go func() {
		for req := range requests {
			switch req.Type {
			case "pty-req", "shell", "window-change", "env":
				if req.WantReply {
					req.Reply(true, nil)
				}
			default:
				if req.WantReply {
					req.Reply(false, nil)
				}
			}
		}
	}()

	sess := &session{
		cfg:      s.cfg,
		username: conn.User(),
		mgt:      mgt,
		ch:       ch,
	}
	sess.run()

	ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{0}))
	io.WriteString(ch, "")
}

// ── Host key ───────────────────────────────────────────────────────────────

func loadOrGenHostKey(cfg Config) (ssh.Signer, error) {
	data, err := os.ReadFile(cfg.HostKeyPath)
	if err != nil {
		if !os.IsNotExist(err) || !cfg.AutoGenHostKey {
			return nil, fmt.Errorf("read host key %s: %w", cfg.HostKeyPath, err)
		}
		logrus.Warnf("gate: host key not found at %s — generating ephemeral ECDSA key", cfg.HostKeyPath)
		return genAndSaveHostKey(cfg.HostKeyPath)
	}
	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("parse host key: %w", err)
	}
	return signer, nil
}

func genAndSaveHostKey(path string) (ssh.Signer, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		return nil, err
	}
	if path != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			logrus.Warnf("gate: mkdir for host key: %v — key will be ephemeral", err)
			return signer, nil
		}
		der, err := x509.MarshalECPrivateKey(key)
		if err != nil {
			return signer, nil
		}
		block := &pem.Block{Type: "EC PRIVATE KEY", Bytes: der}
		if err := os.WriteFile(path, pem.EncodeToMemory(block), 0o600); err != nil {
			logrus.Warnf("gate: save host key: %v — key will be ephemeral", err)
		}
	}
	return signer, nil
}
