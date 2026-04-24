package mysql

import (
	"errors"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"gorm.io/gorm"
)

// neConfigFromCliNe derives a CliNeConfig DTO from CliNe's conf_* fields.
func neConfigFromCliNe(ne *db_models.CliNe) *db_models.CliNeConfig {
	return &db_models.CliNeConfig{
		ID:          ne.ID,
		NeID:        ne.ID,
		IPAddress:   ne.ConfMasterIP,
		Port:        ne.ConfPortMasterSSH,
		Username:    ne.ConfUsername,
		Password:    ne.ConfPassword,
		Protocol:    ne.ConfMode,
		Description: ne.Description,
	}
}

// CreateCliNeConfig writes connection config into the CliNe's conf_* columns.
func (c *Client) CreateCliNeConfig(cfg *db_models.CliNeConfig) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	tx := c.Db.Model(&db_models.CliNe{}).Where("id = ?", cfg.NeID).Updates(map[string]interface{}{
		"conf_master_ip":        cfg.IPAddress,
		"conf_port_master_ssh":  cfg.Port,
		"conf_username":         cfg.Username,
		"conf_password":         cfg.Password,
		"conf_mode":             cfg.Protocol,
		"description":           cfg.Description,
	})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("NE not found")
	}
	return nil
}

// GetCliNeConfigByNeId returns a single-element slice with config derived from CliNe.
func (c *Client) GetCliNeConfigByNeId(neId int64) ([]*db_models.CliNeConfig, error) {
	var ne db_models.CliNe
	tx := c.Db.First(&ne, neId)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return []*db_models.CliNeConfig{}, nil
		}
		return nil, tx.Error
	}
	return []*db_models.CliNeConfig{neConfigFromCliNe(&ne)}, nil
}

// GetCliNeConfigById treats id as the NE id (one config per NE).
func (c *Client) GetCliNeConfigById(id int64) (*db_models.CliNeConfig, error) {
	var ne db_models.CliNe
	tx := c.Db.First(&ne, id)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return neConfigFromCliNe(&ne), nil
}

// UpdateCliNeConfig updates CliNe's conf_* fields using NeID (falls back to ID).
func (c *Client) UpdateCliNeConfig(cfg *db_models.CliNeConfig) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	neId := cfg.NeID
	if neId == 0 {
		neId = cfg.ID
	}
	tx := c.Db.Model(&db_models.CliNe{}).Where("id = ?", neId).Updates(map[string]interface{}{
		"conf_master_ip":        cfg.IPAddress,
		"conf_port_master_ssh":  cfg.Port,
		"conf_username":         cfg.Username,
		"conf_password":         cfg.Password,
		"conf_mode":             cfg.Protocol,
		"description":           cfg.Description,
	})
	return tx.Error
}

// DeleteCliNeConfigById clears the conf_* fields of the CliNe with the given id.
func (c *Client) DeleteCliNeConfigById(id int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	tx := c.Db.Model(&db_models.CliNe{}).Where("id = ?", id).Updates(map[string]interface{}{
		"conf_master_ip":        "",
		"conf_port_master_ssh":  0,
		"conf_username":         "",
		"conf_password":         "",
		"conf_mode":             "",
	})
	return tx.Error
}

// DeleteCliNeConfigByNeId clears the conf_* fields of the CliNe with the given NE id.
func (c *Client) DeleteCliNeConfigByNeId(neId int64) error {
	return c.DeleteCliNeConfigById(neId)
}

// cascade helpers

func (c *Client) DeleteAllUserNeMappingByNeId(neId int64) error {
	tx := c.Db.Where("tbl_ne_id = ?", neId).Delete(&db_models.CliUserNeMapping{})
	if tx == nil {
		return errors.New("no database connection")
	}
	return tx.Error
}

// DeleteAllUserNeMappingByUserId is the user-side counterpart, used by
// PurgeUser when a user is hard-deleted.
func (c *Client) DeleteAllUserNeMappingByUserId(userId int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Where("user_id = ?", userId).Delete(&db_models.CliUserNeMapping{}).Error
}

// DeleteNeMonitorByNeId is a no-op: monitor data is derived from CliNe, no table to delete.
func (c *Client) DeleteNeMonitorByNeId(neId int64) error {
	return nil
}

// DeleteCliNeSlaveByNeId is a no-op: cli_ne_slave table no longer exists.
func (c *Client) DeleteCliNeSlaveByNeId(neId int64) error {
	return nil
}
