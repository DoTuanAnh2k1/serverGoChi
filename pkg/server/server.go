package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/config"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
)

// Server Struct
type Server struct {
	srv *http.Server
	wg  sync.WaitGroup
	cfg config_models.ServerConfig
}

// NewServer Function to Create a New Server Handler
func NewServer(handler http.Handler) *Server {
	cfg := config.GetServerConfig()
	server := &Server{
		srv: &http.Server{
			Addr:    net.JoinHostPort(cfg.Host, cfg.Port),
			Handler: handler,
		},
		cfg: cfg,
	}
	return server
}

// Start Method for Server
func (s *Server) Start() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.wg.Add(1)

	go func() {
		logger.Logger.Infof("Starting server at %s:%s ...", s.cfg.Host, s.cfg.Port)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Errorf("Server error: %v", err)
		}
		s.wg.Done()
	}()
}

// Stop Method for Server
func (s *Server) Stop() {
	timeout := 5 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		if err = s.srv.Close(); err != nil {
			log.Printf("Stop server %s:%s at %v", s.cfg.Host, s.cfg.Port, time.Now().Format(time.RFC3339))
			return
		}
	}
	s.wg.Wait()
}
