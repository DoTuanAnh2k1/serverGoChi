package testutil

import (
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"

	"github.com/sirupsen/logrus"
)

// InitTestLogger sets logger to Error level so tests are not spammed.
func InitTestLogger() {
	l := logrus.New()
	l.SetLevel(logrus.ErrorLevel)
	logger.Logger = l
	logger.DbLogger = l
}
