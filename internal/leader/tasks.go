package leader

import (
	"context"
	"time"

	"go-aa-server/internal/logger"
	"go-aa-server/internal/service"
	"go-aa-server/models/config_models"
)

// RunTasks chạy các tác vụ dành riêng cho leader.
// Hàm này block cho đến khi ctx bị cancel (pod mất lease).
// Khi được gọi lần đầu, nó export ngay một lần để tránh bỏ sót
// nếu pod vừa được khởi động lại sau nửa ngày.
func RunTasks(ctx context.Context, cfg config_models.LeaderConfig) {
	logger.Logger.Infof("leader tasks: started (CSV export dir from CLI_LOG_EXPORT_DIR, daily at %02d:00)",
		cfg.CSVExportHour)

	// Export ngay khi vừa trở thành leader
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

// nextDailyOccurrence trả về khoảng thời gian đến lần kế tiếp của giờ được cấu hình.
func nextDailyOccurrence(hour int) time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return time.Until(next)
}

// nextMonthlyCleanup trả về khoảng thời gian đến 00:00 ngày 1 của tháng tiếp theo.
func nextMonthlyCleanup() time.Duration {
	now := time.Now()
	// Ngày đầu tháng sau, lúc 00:00
	next := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	return time.Until(next)
}
