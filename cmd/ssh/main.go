package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	logrus "github.com/sirupsen/logrus"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/sshcli"
)

func main() {
	_ = godotenv.Load()
	cfg, err := sshcli.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	setLogLevel(cfg.LogLevel)

	srv, err := sshcli.NewServer(cfg)
	if err != nil {
		log.Fatalf("init server: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logrus.Infof("ssh-cli starting — listen=%s mgt=%s ne-config=%s ne-command=%s",
		cfg.ListenAddr, cfg.MgtSvcBase, cfg.NeConfigAddr, cfg.NeCommandAddr)

	if err := srv.ListenAndServe(ctx); err != nil {
		log.Fatalf("serve: %v", err)
	}
	logrus.Info("ssh-cli stopped")
	os.Exit(0)
}

func setLogLevel(s string) {
	lvl, err := logrus.ParseLevel(strings.ToLower(s))
	if err != nil {
		lvl = logrus.InfoLevel
	}
	logrus.SetLevel(lvl)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
}
