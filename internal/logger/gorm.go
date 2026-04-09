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
	SkipErrRecordNotFound bool
	Debug                 bool
}

func NewGormLogger() *GormLogger {
	return &GormLogger{
		Debug:         DbLogger != nil && DbLogger.GetLevel() <= logrus.DebugLevel,
		SlowThreshold: 300 * time.Millisecond,
	}
}

func (l *GormLogger) LogMode(gormlogger.LogLevel) gormlogger.Interface { return l }

func (l *GormLogger) Info(ctx context.Context, s string, args ...interface{}) {
	DbLogger.WithContext(ctx).Infof("db: "+s, args...)
}

func (l *GormLogger) Warn(ctx context.Context, s string, args ...interface{}) {
	DbLogger.WithContext(ctx).Warnf("db: "+s, args...)
}

func (l *GormLogger) Error(ctx context.Context, s string, args ...interface{}) {
	DbLogger.WithContext(ctx).Errorf("db: "+s, args...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	entry := DbLogger.WithContext(ctx).
		WithField("elapsed", elapsed.Round(time.Microsecond).String()).
		WithField("rows", rows)

	if err != nil && !(errors.Is(err, gorm.ErrRecordNotFound) && l.SkipErrRecordNotFound) {
		entry.Errorf("db: query failed: %v — %s", err, sql)
		return
	}
	if l.SlowThreshold > 0 && elapsed > l.SlowThreshold {
		entry.Warnf("db: slow query (>%s) — %s", l.SlowThreshold, sql)
		return
	}
	if l.Debug {
		entry.Debugf("db: %s", sql)
	}
}
