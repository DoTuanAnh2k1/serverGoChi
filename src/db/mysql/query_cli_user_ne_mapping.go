package mysql

import (
	"errors"
	"serverGoChi/models/db_models"
)

func (c *Client) GetCLIUserNeMappingByUserId(userId int64) (*db_models.CliUserNeMapping, error) {
	cond := &db_models.CliUserNeMapping{UserID: userId}
	result := &db_models.CliUserNeMapping{}
	tx := c.Db.First(result, cond)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return result, nil
}
