package mysql

import (
	"errors"
	"serverGoChi/models/db_models"
)

func (c *Client) GetNeListById(id int64) ([]*db_models.CliNe, error) {
	cond := &db_models.CliNe{ID: id}
	var userList []*db_models.CliNe
	tx := c.Db.Find(&userList, cond)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return userList, nil
}

func (c *Client) GetCliNeListBySystemType(systemType string) ([]*db_models.CliNe, error) {
	var cliNeList []*db_models.CliNe
	cond := &db_models.CliNe{SystemType: systemType}
	tx := c.Db.Find(&cliNeList, cond)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return cliNeList, nil
}

func (c *Client) GetCliNeByNeId(id int64) (*db_models.CliNe, error) {
	var cliNe *db_models.CliNe
	cond := &db_models.CliNe{ID: id}
	tx := c.Db.First(&cliNe, cond)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return cliNe, nil
}
