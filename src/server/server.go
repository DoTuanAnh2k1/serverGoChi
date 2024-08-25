package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"serverGoChi/config"
	"serverGoChi/models/config_models"
	"serverGoChi/src/logger"
	"sync"
	"time"
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
	// Initialize New Server
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
	// Initialize Context Handler Without Timeout
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Add to The WaitGroup for The Listener GoRoutine
	// And Wait for 1 Routine to be Done
	s.wg.Add(1)

	// Start The Server

	go func() {
		logger.Logger.Infof("Starting server at %s:%s ...", s.cfg.Host, s.cfg.Port)
		s.srv.ListenAndServe()

		s.wg.Done()
	}()
}

// Stop Method for Server
func (s *Server) Stop() {
	// Initialize Timeout
	timeout := 5 * time.Second

	// Initialize Context Handler With Timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Hanlde Any Error While Stopping Server
	if err := s.srv.Shutdown(ctx); err != nil {
		if err = s.srv.Close(); err != nil {
			log.Printf("Stop server %s:%s at %v", s.cfg.Host, s.cfg.Port, time.Now().Format(time.RFC3339))
			return
		}
	}
	s.wg.Wait()
}
