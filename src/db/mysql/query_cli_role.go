package mysql

import (
	"errors"
	"gorm.io/gorm"
	"serverGoChi/models/db_models"
)

func (c *Client) GetCliRole(cliRole db_models.CliRole) (*db_models.CliRole, error) {
	cond := &cliRole
	result := &db_models.CliRole{}
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

func (c *Client) CreateCliRole(cliRole db_models.CliRole) error {
	cond := &cliRole
	tx := c.Db.Create(cond)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (c *Client) DeleteCliRole(cliRole db_models.CliRole) error {
	cond := &cliRole
	tx := c.Db.Delete(cond)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (c *Client) GetAllCliRole() ([]*db_models.CliRole, error) {
	var cliRoleList []*db_models.CliRole
	tx := c.Db.Find(&cliRoleList)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return cliRoleList, nil
}
