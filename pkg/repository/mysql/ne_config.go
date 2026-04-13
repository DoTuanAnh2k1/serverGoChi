package mysql

import (
	"errors"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"gorm.io/gorm"
)

func (c *Client) CreateCliNeConfig(cfg *db_models.CliNeConfig) error {
	tx := c.Db.Create(cfg)
	if tx == nil {
		return errors.New("no database connection")
	}
	return tx.Error
}

func (c *Client) GetCliNeConfigByNeId(neId int64) ([]*db_models.CliNeConfig, error) {
	var list []*db_models.CliNeConfig
	tx := c.Db.Where("ne_id = ?", neId).Find(&list)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	return list, tx.Error
}

func (c *Client) GetCliNeConfigById(id int64) (*db_models.CliNeConfig, error) {
	result := &db_models.CliNeConfig{}
	tx := c.Db.First(result, id)
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

func (c *Client) UpdateCliNeConfig(cfg *db_models.CliNeConfig) error {
	tx := c.Db.Save(cfg)
	if tx == nil {
		return errors.New("no database connection")
	}
	return tx.Error
}

func (c *Client) DeleteCliNeConfigById(id int64) error {
	tx := c.Db.Delete(&db_models.CliNeConfig{}, id)
	if tx == nil {
		return errors.New("no database connection")
	}
	return tx.Error
}

func (c *Client) DeleteCliNeConfigByNeId(neId int64) error {
	tx := c.Db.Where("ne_id = ?", neId).Delete(&db_models.CliNeConfig{})
	if tx == nil {
		return errors.New("no database connection")
	}
	return tx.Error
}

// cascade helpers

func (c *Client) DeleteAllUserNeMappingByNeId(neId int64) error {
	tx := c.Db.Where("tbl_ne_id = ?", neId).Delete(&db_models.CliUserNeMapping{})
	if tx == nil {
		return errors.New("no database connection")
	}
	return tx.Error
}

func (c *Client) DeleteNeMonitorByNeId(neId int64) error {
	tx := c.Db.Where("ne_id = ?", neId).Delete(&db_models.CliNeMonitor{})
	if tx == nil {
		return errors.New("no database connection")
	}
	return tx.Error
}

func (c *Client) DeleteCliNeSlaveByNeId(neId int64) error {
	tx := c.Db.Where("ne_id = ?", neId).Delete(&db_models.CliNeSlave{})
	if tx == nil {
		return errors.New("no database connection")
	}
	return tx.Error
}
