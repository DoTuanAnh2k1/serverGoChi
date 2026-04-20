package postgres

import (
	"errors"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"gorm.io/gorm"
)

// ── cli_group CRUD ───────────────────────────────────────────────────────────

func (c *Client) CreateGroup(g *db_models.CliGroup) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Create(g).Error
}

func (c *Client) GetGroupById(id int64) (*db_models.CliGroup, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var g db_models.CliGroup
	tx := c.Db.First(&g, id)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &g, nil
}

func (c *Client) GetGroupByName(name string) (*db_models.CliGroup, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var g db_models.CliGroup
	tx := c.Db.First(&g, &db_models.CliGroup{Name: name})
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &g, nil
}

func (c *Client) GetAllGroups() ([]*db_models.CliGroup, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var list []*db_models.CliGroup
	return list, c.Db.Find(&list).Error
}

func (c *Client) UpdateGroup(g *db_models.CliGroup) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Save(g).Error
}

func (c *Client) DeleteGroupById(id int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Delete(&db_models.CliGroup{}, id).Error
}

// ── cli_user_group_mapping ───────────────────────────────────────────────────

func (c *Client) CreateUserGroupMapping(m *db_models.CliUserGroupMapping) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Create(m).Error
}

func (c *Client) DeleteUserGroupMapping(m *db_models.CliUserGroupMapping) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Delete(m).Error
}

func (c *Client) GetAllGroupsOfUser(userId int64) ([]*db_models.CliUserGroupMapping, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var list []*db_models.CliUserGroupMapping
	return list, c.Db.Find(&list, &db_models.CliUserGroupMapping{UserID: userId}).Error
}

func (c *Client) GetAllUsersOfGroup(groupId int64) ([]*db_models.CliUserGroupMapping, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var list []*db_models.CliUserGroupMapping
	return list, c.Db.Find(&list, &db_models.CliUserGroupMapping{GroupID: groupId}).Error
}

func (c *Client) DeleteAllUserGroupMappingByUserId(userId int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Where("user_id = ?", userId).Delete(&db_models.CliUserGroupMapping{}).Error
}

func (c *Client) DeleteAllUserGroupMappingByGroupId(groupId int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Where("group_id = ?", groupId).Delete(&db_models.CliUserGroupMapping{}).Error
}

// ── cli_group_ne_mapping ─────────────────────────────────────────────────────

func (c *Client) CreateGroupNeMapping(m *db_models.CliGroupNeMapping) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Create(m).Error
}

func (c *Client) DeleteGroupNeMapping(m *db_models.CliGroupNeMapping) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Delete(m).Error
}

func (c *Client) GetAllNesOfGroup(groupId int64) ([]*db_models.CliGroupNeMapping, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var list []*db_models.CliGroupNeMapping
	return list, c.Db.Find(&list, &db_models.CliGroupNeMapping{GroupID: groupId}).Error
}

func (c *Client) GetAllGroupsOfNe(neId int64) ([]*db_models.CliGroupNeMapping, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var list []*db_models.CliGroupNeMapping
	return list, c.Db.Find(&list, &db_models.CliGroupNeMapping{TblNeID: neId}).Error
}

func (c *Client) DeleteAllGroupNeMappingByGroupId(groupId int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Where("group_id = ?", groupId).Delete(&db_models.CliGroupNeMapping{}).Error
}

func (c *Client) DeleteAllGroupNeMappingByNeId(neId int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Where("tbl_ne_id = ?", neId).Delete(&db_models.CliGroupNeMapping{}).Error
}
