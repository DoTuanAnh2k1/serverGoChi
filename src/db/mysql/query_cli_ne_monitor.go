package mysql

import (
	"errors"
	"gorm.io/gorm"
	"serverGoChi/models/db_models"
)

func (c *Client) GetNeMonitorById(id int64) (*db_models.CliNeMonitor, error) {
	cond := &db_models.CliNeMonitor{NeID: id}
	result := &db_models.CliNeMonitor{}
	tx := c.Db.First(result, cond)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return result, nil
}
