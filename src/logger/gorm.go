package logger

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type GormLogger struct {
	SlowThreshold         time.Duration
	SourceField           string
	SkipErrRecordNotFound bool
	Debug                 bool
}

func NewGormLogger() *GormLogger {
	debug := false
	if DbLogger.GetLevel() <= logrus.DebugLevel {
		debug = true
	}
	return &GormLogger{
		Debug: debug,
	}
}

func (l *GormLogger) LogMode(gormlogger.LogLevel) gormlogger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, s string, args ...interface{}) {
	DbLogger.WithContext(ctx).WithField("TAG", "DB").Infof(s, args...)
}

func (l *GormLogger) Warn(ctx context.Context, s string, args ...interface{}) {
	DbLogger.WithContext(ctx).WithField("TAG", "DB").Warnf(s, args...)
}

func (l *GormLogger) Error(ctx context.Context, s string, args ...interface{}) {
	DbLogger.WithContext(ctx).WithField("TAG", "DB").Errorf(s, args...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, _ := fc()

	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
		DbLogger.WithContext(ctx).WithField("TAG", "DB").Errorf("%s [%s]", sql, elapsed)
		return
	}

	if l.SlowThreshold != 0 && elapsed > l.SlowThreshold {
		DbLogger.WithContext(ctx).WithField("TAG", "DB").Warnf("%s [%s]", sql, elapsed)
		return
	}

	if l.Debug {
		DbLogger.WithContext(ctx).WithField("TAG", "DB").Debugf("%s [%s]", sql, elapsed)
	}
}
