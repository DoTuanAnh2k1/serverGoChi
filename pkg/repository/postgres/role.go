package postgres

import (
	"errors"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"

	"gorm.io/gorm"
)

func (c *Client) GetCliRole(cliRole *db_models.CliRole) (*db_models.CliRole, error) {
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

func (c *Client) CreateCliRole(cliRole *db_models.CliRole) error {
	tx := c.Db.Create(&cliRole)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (c *Client) DeleteCliRole(cliRole *db_models.CliRole) error {
	tx := c.Db.Delete(cliRole, cliRole.RoleID)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (c *Client) GetAllCliRole() ([]*db_models.CliRole, error) {
	var list []*db_models.CliRole
	tx := c.Db.Find(&list)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return list, nil
}

func (c *Client) GetRolesById(id int64) ([]*db_models.CliRoleUserMapping, error) {
	cond := &db_models.CliRoleUserMapping{UserID: id}
	var list []*db_models.CliRoleUserMapping
	tx := c.Db.Find(&list, cond)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return list, nil
}

func (c *Client) AddRole(role *db_models.CliRoleUserMapping) error {
	tx := c.Db.Save(role)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (c *Client) DeleteRole(role *db_models.CliRoleUserMapping) error {
	tx := c.Db.Delete(&role)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}
