package service

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// DeleteOldHistory deletes history older than the previous month.
func DeleteOldHistory(now time.Time) (int64, error) {
	// Cutoff = first day of previous month
	cutoff := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	deleted, err := store.GetSingleton().DeleteHistoryBefore(cutoff)
	if err != nil {
		logger.Logger.Errorf("DeleteOldHistory: failed (cutoff=%s): %v", cutoff.Format("2006-01-02"), err)
		return 0, err
	}
	logger.Logger.Infof("DeleteOldHistory: removed %d records older than %s", deleted, cutoff.Format("2006-01-02"))
	return deleted, nil
}

// GetRecentHistory returns the N most recent history records.
func GetRecentHistory(limit int) ([]db_models.CliOperationHistory, error) {
	return store.GetSingleton().GetRecentHistory(limit)
}

func GetRecentHistoryFiltered(limit int, scope, neName, account string) ([]db_models.CliOperationHistory, error) {
	return store.GetSingleton().GetRecentHistoryFiltered(limit, scope, neName, account)
}

// SaveHistoryCommand saves a history record to the database.
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

// ExportDailyHistoryByNe exports daily history grouped by NE to CSV files.
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

	// Group by NE name
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

// sanitizeFilename replaces invalid filename characters.
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
