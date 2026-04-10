package service

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/internal/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/store"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// DeleteOldHistory xoá toàn bộ lịch sử lệnh cũ hơn tháng trước.
// Ví dụ: nếu now là tháng 4/2026 thì cutoff = 2026-03-01 00:00:00,
// tức là chỉ giữ lại dữ liệu từ tháng 3/2026 trở đi.
func DeleteOldHistory(now time.Time) (int64, error) {
	// Cutoff = ngày đầu tiên của tháng trước
	cutoff := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	deleted, err := store.GetSingleton().DeleteHistoryBefore(cutoff)
	if err != nil {
		logger.Logger.Errorf("DeleteOldHistory: failed (cutoff=%s): %v", cutoff.Format("2006-01-02"), err)
		return 0, err
	}
	logger.Logger.Infof("DeleteOldHistory: removed %d records older than %s", deleted, cutoff.Format("2006-01-02"))
	return deleted, nil
}

// SaveHistoryCommand lưu một bản ghi lịch sử lệnh vào database.
func SaveHistoryCommand(historyCommand db_models.CliOperationHistory) error {
	sto := store.GetSingleton()
	logger.Logger.Debug("Save command")
	err := sto.SaveHistoryCommand(historyCommand)
	if err != nil {
		logger.Logger.Error("Cant save history command: ", err)
		return err
	}
	return nil
}

// ExportDailyHistoryByNe lấy toàn bộ lịch sử lệnh trong ngày `date`,
// nhóm theo tên NE và ghi ra các file CSV.
// Thư mục xuất đọc từ biến môi trường CLI_LOG_EXPORT_DIR.
// Tên file: cli_log_<ne_site>_YYYYMMDD.csv
func ExportDailyHistoryByNe(date time.Time) error {
	exportDir := os.Getenv("CLI_LOG_EXPORT_DIR")
	if exportDir == "" {
		exportDir = "."
	}

	histories, err := store.GetSingleton().GetDailyOperationHistory(date)
	if err != nil {
		return fmt.Errorf("query daily history: %w", err)
	}

	if len(histories) == 0 {
		logger.Logger.Infof("export: no history records found for %s", date.Format("2006-01-02"))
		return nil
	}

	// Nhóm theo NE name
	grouped := make(map[string][]db_models.CliOperationHistory)
	for _, h := range histories {
		key := h.NeName
		if strings.TrimSpace(key) == "" {
			key = "unknown"
		}
		grouped[key] = append(grouped[key], h)
	}

	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return fmt.Errorf("create export dir %q: %w", exportDir, err)
	}

	dateStr := date.Format("20060102")
	var writeErrors []string
	for neName, records := range grouped {
		filename := fmt.Sprintf("cli_log_%s_%s.csv", sanitizeFilename(neName), dateStr)
		filePath := filepath.Join(exportDir, filename)
		if err := writeCSV(filePath, records); err != nil {
			writeErrors = append(writeErrors, fmt.Sprintf("NE %q: %v", neName, err))
		} else {
			logger.Logger.Infof("export: wrote %d records to %s", len(records), filePath)
		}
	}

	if len(writeErrors) > 0 {
		return fmt.Errorf("some CSV files failed to write: %s", strings.Join(writeErrors, "; "))
	}
	return nil
}

var csvHeader = []string{
	"user", "command", "result", "execute_time", "ne", "remote_address", "id",
}

func writeCSV(path string, records []db_models.CliOperationHistory) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	if err := w.Write(csvHeader); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for _, r := range records {
		row := []string{
			r.Account,
			r.CmdName,
			r.Result,
			r.ExecutedTime.Format(time.RFC3339),
			r.NeName,
			r.IPAddress,
			fmt.Sprintf("%d", r.ID),
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("write row id=%d: %w", r.ID, err)
		}
	}

	return w.Error()
}

// sanitizeFilename thay thế các ký tự không hợp lệ trong tên file.
func sanitizeFilename(name string) string {
	invalid := `/\:*?"<>|`
	result := make([]byte, len(name))
	for i := 0; i < len(name); i++ {
		if strings.ContainsRune(invalid, rune(name[i])) {
			result[i] = '_'
		} else {
			result[i] = name[i]
		}
	}
	return string(result)
}
