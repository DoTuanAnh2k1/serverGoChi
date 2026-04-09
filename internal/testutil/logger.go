package testutil

import (
	"go-aa-server/internal/logger"

	"github.com/sirupsen/logrus"
)

// InitTestLogger khởi tạo logger ở mức Error để test không bị spam log
// nhưng vẫn thấy các lỗi thực sự. Gọi trong TestMain của từng package test.
func InitTestLogger() {
	l := logrus.New()
	l.SetLevel(logrus.ErrorLevel)
	logger.Logger = l
	logger.DbLogger = l
}
