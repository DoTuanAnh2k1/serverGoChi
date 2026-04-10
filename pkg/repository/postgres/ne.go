package postgres

import (
	"errors"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"

	"gorm.io/gorm"
)

func (c *Client) GetNeListById(id int64) ([]*db_models.CliNe, error) {
	cond := &db_models.CliNe{ID: id}
	var list []*db_models.CliNe
	tx := c.Db.Find(&list, cond)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return list, nil
}

func (c *Client) GetCliNeListBySystemType(systemType string) ([]*db_models.CliNe, error) {
	var list []*db_models.CliNe
	cond := &db_models.CliNe{SystemType: systemType}
	tx := c.Db.Find(&list, cond)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return list, nil
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

func (c *Client) CreateUserNeMapping(m *db_models.CliUserNeMapping) error {
	tx := c.Db.Create(m)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (c *Client) DeleteUserNeMapping(m *db_models.CliUserNeMapping) error {
	tx := c.Db.Delete(m)
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
