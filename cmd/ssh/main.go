package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/sshcli"
)

func main() {
	_ = godotenv.Load()
	cfg, err := sshcli.LoadConfig()
	if err != nil {
		log.Fatalf("gate: load config: %v", err)
	}

	lvl, err := logrus.ParseLevel(strings.ToLower(cfg.LogLevel))
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logrus.SetLevel(lvl)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	srv, err := sshcli.NewServer(cfg)
	if err != nil {
		log.Fatalf("gate: init server: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logrus.Infof("cli-gate starting — listen=%s mgt=%s", cfg.ListenAddr, cfg.MgtSvcBase)

	if err := srv.ListenAndServe(ctx); err != nil {
		log.Fatalf("gate: serve: %v", err)
	}

	logrus.Info("cli-gate stopped")
	os.Exit(0)
}
