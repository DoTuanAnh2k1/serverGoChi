package leader

import (
	"context"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
)

// RunTasks runs leader-only tasks. Blocks until ctx is cancelled.
// Exports immediately on first run to avoid gaps.
func RunTasks(ctx context.Context, cfg config_models.LeaderConfig) {
	logger.Logger.Infof("leader tasks: started (CSV export dir from CLI_LOG_EXPORT_DIR, daily at %02d:00)",
		cfg.CSVExportHour)

	// Export immediately on becoming leader
	runCSVExport()

	for {
		csvWait := nextDailyOccurrence(cfg.CSVExportHour)
		cleanupWait := nextMonthlyCleanup()

		logger.Logger.Infof("leader tasks: next CSV export in %s | next history cleanup in %s",
			csvWait.Round(time.Second), cleanupWait.Round(time.Second))

		select {
		case <-ctx.Done():
			logger.Logger.Info("leader tasks: context cancelled — stopping")
			return
		case <-time.After(csvWait):
			runCSVExport()
		case <-time.After(cleanupWait):
			runHistoryCleanup()
		}
	}
}

func runCSVExport() {
	date := time.Now()
	logger.Logger.Infof("leader tasks: exporting daily history for %s", date.Format("2006-01-02"))

	if err := service.ExportDailyHistoryByNe(date); err != nil {
		logger.Logger.Errorf("leader tasks: CSV export failed: %v", err)
		return
	}

	logger.Logger.Infof("leader tasks: CSV export completed for %s", date.Format("2006-01-02"))
}

func runHistoryCleanup() {
	runHistoryCleanupAt(time.Now())
}

func runHistoryCleanupAt(now time.Time) {
	logger.Logger.Infof("leader tasks: running monthly history cleanup (now=%s)", now.Format("2006-01"))

	deleted, err := service.DeleteOldHistory(now)
	if err != nil {
		logger.Logger.Errorf("leader tasks: history cleanup failed: %v", err)
		return
	}

	logger.Logger.Infof("leader tasks: history cleanup done, removed %d records", deleted)
}

// nextDailyOccurrence returns duration until the next configured hour.
func nextDailyOccurrence(hour int) time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return time.Until(next)
}

// nextMonthlyCleanup returns duration until 00:00 on the 1st of next month.
func nextMonthlyCleanup() time.Duration {
	now := time.Now()
	// First day of next month at 00:00
	next := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	return time.Until(next)
}
