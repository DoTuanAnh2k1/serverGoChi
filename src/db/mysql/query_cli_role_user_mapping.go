package mysql

import (
	"errors"
	"serverGoChi/models/db_models"
)

func (c *Client) GetRolesById(id int64) ([]*db_models.CliRoleUserMapping, error) {
	cond := &db_models.CliRoleUserMapping{
		UserID: id,
	}
	var roleList []*db_models.CliRoleUserMapping
	tx := c.Db.Find(&roleList, cond)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return roleList, nil
}
