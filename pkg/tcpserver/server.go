// Package tcpserver provides a simple TCP server that reads lines per connection
// and writes received data to list_subscribers_results.<index> files.
package tcpserver

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
)

// Server listens on TCP and saves each connection's data to a file.
type Server struct {
	addr    string
	dataDir string

	mu       sync.Mutex   // protects file index selection
	listener net.Listener
}

// New creates a new Server. addr e.g. ":3675", dataDir is the output directory.
func New(addr, dataDir string) *Server {
	return &Server{addr: addr, dataDir: dataDir}
}

// Start begins listening in a separate goroutine.
func (s *Server) Start() error {
	if err := os.MkdirAll(s.dataDir, 0755); err != nil {
		return fmt.Errorf("tcp: create data dir %q: %w", s.dataDir, err)
	}

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("tcp: listen %s: %w", s.addr, err)
	}
	s.listener = ln

	logger.Logger.
		WithField("addr", s.addr).
		WithField("data_dir", s.dataDir).
		Info("tcp: server started")

	go s.acceptLoop(ln)
	return nil
}

// Stop closes the listener; open connections will finish naturally.
func (s *Server) Stop() {
	if s.listener != nil {
		_ = s.listener.Close()
		logger.Logger.Info("tcp: server stopped")
	}
}

func (s *Server) acceptLoop(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			// net.ErrClosed occurs when Stop() is called — not a real error
			if isClosedErr(err) {
				return
			}
			logger.Logger.Errorf("tcp: accept: %v", err)
			return
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	remote := conn.RemoteAddr().String()
	log := logger.Logger.WithField("remote", remote)
	log.Info("tcp: connection opened")

	var lines []string
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Errorf("tcp: read error: %v", err)
	}

	log.WithField("lines", len(lines)).Info("tcp: connection closed")

	if len(lines) == 0 {
		log.Debug("tcp: no data received — skipping file write")
		return
	}

	path, err := s.writeFile(lines)
	if err != nil {
		log.Errorf("tcp: save data: %v", err)
		return
	}
	log.WithField("file", path).WithField("lines", len(lines)).Infof("tcp: saved %d lines → %s", len(lines), path)
}

// writeFile picks the next index and writes data to file.
// Mutex ensures concurrent connections don't write to the same file.
func (s *Server) writeFile(lines []string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	path, err := s.nextAvailablePath()
	if err != nil {
		return "", err
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("create file %q: %w", path, err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return "", fmt.Errorf("write to %q: %w", path, err)
		}
	}
	if err := w.Flush(); err != nil {
		return "", fmt.Errorf("flush %q: %w", path, err)
	}
	return path, nil
}

// nextAvailablePath finds the smallest unused index.
// Must be called while holding s.mu.
func (s *Server) nextAvailablePath() (string, error) {
	for idx := 0; ; idx++ {
		path := filepath.Join(s.dataDir, fmt.Sprintf("list_subscribers_results.%d", idx))
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path, nil
		} else if err != nil {
			return "", fmt.Errorf("stat %q: %w", path, err)
		}
	}
}

func isClosedErr(err error) bool {
	if err == nil {
		return false
	}
	// net package does not export ErrClosed directly; use string match for compatibility
	return contains(err.Error(), "use of closed network connection")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
