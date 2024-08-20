package mysql

import (
	"errors"
	"serverGoChi/models/db_models"
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
