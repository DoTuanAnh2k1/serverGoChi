package handler_test

import (
	"os"
	"testing"

	"github.com/DoTuanAnh2k1/serverGoChi/internal/config"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	l := logrus.New()
	l.SetLevel(logrus.ErrorLevel)
	logger.Logger = l
	logger.DbLogger = l

	config.Init(&config_models.Config{
		Token: config_models.TokenConfig{
			SecretKey:   "test-secret-key-for-handler-tests",
			ExpiryHours: 1,
		},
	})
	os.Exit(m.Run())
}
