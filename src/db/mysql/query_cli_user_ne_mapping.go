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

func (c *Client) CreateUserNeMapping(cliUserMapping *db_models.CliUserNeMapping) error {
	tx := c.Db.Create(cliUserMapping)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (c *Client) DeleteUserNeMapping(cliUserMapping *db_models.CliUserNeMapping) error {
	tx := c.Db.Delete(cliUserMapping)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (c *Client) GetAllNeOfUserByUserId(userId int64) ([]*db_models.CliUserNeMapping, error) {
	cond := &db_models.CliUserNeMapping{UserID: userId}
	var result []*db_models.CliUserNeMapping
	tx := c.Db.Find(&result, cond)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return result, nil
}
