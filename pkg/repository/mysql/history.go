package mysql

import (
	"errors"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

func (c *Client) SaveHistoryCommand(history db_models.CliOperationHistory) error {
	cond := &history
	tx := c.Db.Create(cond)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

// GetRecentHistory returns the N most recent history records.
func (c *Client) GetRecentHistory(limit int) ([]db_models.CliOperationHistory, error) {
	var records []db_models.CliOperationHistory
	tx := c.Db.Order("created_date DESC").Limit(limit).Find(&records)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return records, nil
}

// DeleteHistoryBefore deletes all cli_operation_history records
// with created_date < cutoff. Returns deleted count.
func (c *Client) DeleteHistoryBefore(cutoff time.Time) (int64, error) {
	tx := c.Db.
		Where("created_date < ?", cutoff).
		Delete(&db_models.CliOperationHistory{})
	if tx.Error != nil {
		return 0, tx.Error
	}
	return tx.RowsAffected, nil
}

// GetDailyOperationHistory returns all cli_operation_history records
// for the given date (00:00:00 to 23:59:59 local time).
func (c *Client) GetDailyOperationHistory(date time.Time) ([]db_models.CliOperationHistory, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)

	var records []db_models.CliOperationHistory
	tx := c.Db.
		Where("created_date >= ? AND created_date < ?", start, end).
		Order("ne_name ASC, created_date ASC").
		Find(&records)

	if tx.Error != nil {
		return nil, tx.Error
	}
	return records, nil
}
