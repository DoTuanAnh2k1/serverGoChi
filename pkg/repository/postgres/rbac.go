package postgres

import (
	"errors"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"gorm.io/gorm"
)

// ───────── cli_ne_profile ─────────

func (c *Client) CreateNeProfile(p *db_models.CliNeProfile) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Create(p).Error
}

func (c *Client) GetNeProfileById(id int64) (*db_models.CliNeProfile, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var p db_models.CliNeProfile
	tx := c.Db.First(&p, id)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &p, nil
}

func (c *Client) GetNeProfileByName(name string) (*db_models.CliNeProfile, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var p db_models.CliNeProfile
	tx := c.Db.Where("name = ?", name).First(&p)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &p, nil
}

func (c *Client) ListNeProfiles() ([]*db_models.CliNeProfile, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var out []*db_models.CliNeProfile
	return out, c.Db.Find(&out).Error
}

func (c *Client) UpdateNeProfile(p *db_models.CliNeProfile) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Save(p).Error
}

func (c *Client) DeleteNeProfileById(id int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Delete(&db_models.CliNeProfile{}, id).Error
}

// ───────── cli_command_def ─────────

func (c *Client) CreateCommandDef(d *db_models.CliCommandDef) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Create(d).Error
}

func (c *Client) GetCommandDefById(id int64) (*db_models.CliCommandDef, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var d db_models.CliCommandDef
	tx := c.Db.First(&d, id)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &d, nil
}

func (c *Client) ListCommandDefs(service, neProfile, category string) ([]*db_models.CliCommandDef, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	q := c.Db.Model(&db_models.CliCommandDef{})
	if service != "" {
		q = q.Where("service = ?", service)
	}
	if neProfile != "" {
		q = q.Where("ne_profile = ?", neProfile)
	}
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var out []*db_models.CliCommandDef
	return out, q.Find(&out).Error
}

func (c *Client) UpdateCommandDef(d *db_models.CliCommandDef) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Save(d).Error
}

func (c *Client) DeleteCommandDefById(id int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("command_def_id = ?", id).Delete(&db_models.CliCommandGroupMapping{}).Error; err != nil {
			return err
		}
		return tx.Delete(&db_models.CliCommandDef{}, id).Error
	})
}

// ───────── cli_command_group ─────────

func (c *Client) CreateCommandGroup(g *db_models.CliCommandGroup) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Create(g).Error
}

func (c *Client) GetCommandGroupById(id int64) (*db_models.CliCommandGroup, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var g db_models.CliCommandGroup
	tx := c.Db.First(&g, id)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &g, nil
}

func (c *Client) GetCommandGroupByName(name string) (*db_models.CliCommandGroup, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var g db_models.CliCommandGroup
	tx := c.Db.Where("name = ?", name).First(&g)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &g, nil
}

func (c *Client) ListCommandGroups(service, neProfile string) ([]*db_models.CliCommandGroup, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	q := c.Db.Model(&db_models.CliCommandGroup{})
	if service != "" {
		q = q.Where("service = ?", service)
	}
	if neProfile != "" {
		q = q.Where("ne_profile = ?", neProfile)
	}
	var out []*db_models.CliCommandGroup
	return out, q.Find(&out).Error
}

func (c *Client) UpdateCommandGroup(g *db_models.CliCommandGroup) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Save(g).Error
}

func (c *Client) DeleteCommandGroupById(id int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("command_group_id = ?", id).Delete(&db_models.CliCommandGroupMapping{}).Error; err != nil {
			return err
		}
		return tx.Delete(&db_models.CliCommandGroup{}, id).Error
	})
}

// ───────── cli_command_group_mapping ─────────

func (c *Client) AddCommandToGroup(m *db_models.CliCommandGroupMapping) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Create(m).Error
}

func (c *Client) RemoveCommandFromGroup(m *db_models.CliCommandGroupMapping) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Where("command_group_id = ? AND command_def_id = ?", m.CommandGroupID, m.CommandDefID).
		Delete(&db_models.CliCommandGroupMapping{}).Error
}

func (c *Client) ListCommandsOfGroup(groupId int64) ([]*db_models.CliCommandDef, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var defs []*db_models.CliCommandDef
	err := c.Db.Table(db_models.TableNameCliCommandDef+" AS d").
		Joins("INNER JOIN "+db_models.TableNameCliCommandGroupMapping+" AS m ON m.command_def_id = d.id").
		Where("m.command_group_id = ?", groupId).
		Find(&defs).Error
	return defs, err
}

func (c *Client) ListGroupsOfCommand(commandId int64) ([]*db_models.CliCommandGroup, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var groups []*db_models.CliCommandGroup
	err := c.Db.Table(db_models.TableNameCliCommandGroup+" AS g").
		Joins("INNER JOIN "+db_models.TableNameCliCommandGroupMapping+" AS m ON m.command_group_id = g.id").
		Where("m.command_def_id = ?", commandId).
		Find(&groups).Error
	return groups, err
}

func (c *Client) DeleteAllCommandGroupMappingByGroupId(groupId int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Where("command_group_id = ?", groupId).Delete(&db_models.CliCommandGroupMapping{}).Error
}

func (c *Client) DeleteAllCommandGroupMappingByCommandId(commandId int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Where("command_def_id = ?", commandId).Delete(&db_models.CliCommandGroupMapping{}).Error
}

// ───────── cli_group_cmd_permission ─────────

func (c *Client) CreateGroupCmdPermission(p *db_models.CliGroupCmdPermission) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Create(p).Error
}

func (c *Client) GetGroupCmdPermissionById(id int64) (*db_models.CliGroupCmdPermission, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var p db_models.CliGroupCmdPermission
	tx := c.Db.First(&p, id)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &p, nil
}

func (c *Client) ListGroupCmdPermissions(groupId int64) ([]*db_models.CliGroupCmdPermission, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var out []*db_models.CliGroupCmdPermission
	return out, c.Db.Where("group_id = ?", groupId).Find(&out).Error
}

func (c *Client) DeleteGroupCmdPermissionById(id int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Delete(&db_models.CliGroupCmdPermission{}, id).Error
}

func (c *Client) DeleteAllGroupCmdPermissionByGroupId(groupId int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Where("group_id = ?", groupId).Delete(&db_models.CliGroupCmdPermission{}).Error
}
