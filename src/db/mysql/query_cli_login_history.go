package mysql

import (
	"errors"
	"serverGoChi/models/db_models"
	"time"
)

func (c *Client) UpdateLoginHistory(username, ipAddress string, timeLogin time.Time) error {
	cond := &db_models.CliLoginHistory{
		UserName:  username,
		IPAddress: ipAddress,
		TimeLogin: timeLogin,
	}
	tx := c.Db.Create(cond)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}
