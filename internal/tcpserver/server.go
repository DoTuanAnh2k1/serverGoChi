// Package tcpserver cung cấp một TCP server đơn giản:
// mỗi kết nối được đọc theo từng dòng cho đến khi client ngắt,
// sau đó toàn bộ dữ liệu nhận được ghi vào file list_subscribers_results.<index>.
package tcpserver

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"go-aa-server/internal/logger"
)

// Server lắng nghe TCP và lưu dữ liệu từ mỗi kết nối ra file.
type Server struct {
	addr    string
	dataDir string

	mu       sync.Mutex   // bảo vệ việc chọn index file
	listener net.Listener
}

// New tạo Server mới. addr ví dụ: ":3675", dataDir là thư mục lưu file.
func New(addr, dataDir string) *Server {
	return &Server{addr: addr, dataDir: dataDir}
}

// Start bắt đầu lắng nghe trong goroutine riêng.
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

// Stop đóng listener, các kết nối đang mở sẽ kết thúc tự nhiên.
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
			// net.ErrClosed xảy ra khi Stop() được gọi — không phải lỗi thực sự
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

// writeFile lựa chọn index tiếp theo và ghi dữ liệu ra file.
// Mutex đảm bảo hai kết nối đóng cùng lúc không ghi vào cùng file.
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

// nextAvailablePath tìm index nhỏ nhất chưa được dùng.
// Phải gọi trong khi giữ s.mu.
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
	// net package không export ErrClosed trực tiếp, kiểm tra qua string
	// (Go 1.16+ có net.ErrClosed nhưng dùng string match để tương thích)
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
