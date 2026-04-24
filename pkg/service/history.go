package service

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

// DeleteOldHistory deletes history older than the first day of the previous
// month. Returns number of rows removed.
func DeleteOldHistory(now time.Time) (int64, error) {
	cutoff := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	deleted, err := store.GetSingleton().DeleteHistoryBefore(cutoff)
	if err != nil {
		logger.Logger.Errorf("DeleteOldHistory: failed (cutoff=%s): %v", cutoff.Format("2006-01-02"), err)
		return 0, err
	}
	logger.Logger.Infof("DeleteOldHistory: removed %d records older than %s", deleted, cutoff.Format("2006-01-02"))
	return deleted, nil
}

func GetRecentHistory(limit int) ([]db_models.OperationHistory, error) {
	return store.GetSingleton().GetRecentHistory(limit)
}

func GetRecentHistoryFiltered(limit int, scope, neNamespace, account string) ([]db_models.OperationHistory, error) {
	return store.GetSingleton().GetRecentHistoryFiltered(limit, scope, neNamespace, account)
}

func SaveHistory(h db_models.OperationHistory) error {
	if err := store.GetSingleton().SaveOperationHistory(h); err != nil {
		logger.Logger.Errorf("SaveHistory: %v", err)
		return err
	}
	return nil
}

// ExportDailyHistoryByNe writes one CSV per NE for the given date under
// CLI_LOG_EXPORT_DIR (defaults to CWD).
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

	grouped := make(map[string][]db_models.OperationHistory)
	for _, h := range histories {
		key := h.NeNamespace
		if strings.TrimSpace(key) == "" {
			key = "unknown"
		}
		grouped[key] = append(grouped[key], h)
	}

	if err := os.MkdirAll(exportDir, 0o755); err != nil {
		return fmt.Errorf("create export dir %q: %w", exportDir, err)
	}

	dateStr := date.Format("20060102")
	var writeErrors []string
	for ns, records := range grouped {
		filename := fmt.Sprintf("cli_log_%s_%s.csv", sanitizeFilename(ns), dateStr)
		filePath := filepath.Join(exportDir, filename)
		if err := writeCSV(filePath, records); err != nil {
			writeErrors = append(writeErrors, fmt.Sprintf("NE %q: %v", ns, err))
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
	"account", "cmd_text", "result", "executed_time", "ne_namespace", "ne_ip", "ip_address", "scope", "id",
}

func writeCSV(path string, records []db_models.OperationHistory) error {
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
			r.CmdText,
			r.Result,
			r.ExecutedTime.Format(time.RFC3339),
			r.NeNamespace,
			r.NeIP,
			r.IPAddress,
			r.Scope,
			fmt.Sprintf("%d", r.ID),
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("write row id=%d: %w", r.ID, err)
		}
	}
	return w.Error()
}

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
