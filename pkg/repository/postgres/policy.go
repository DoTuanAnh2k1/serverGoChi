package postgres

import (
	"errors"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"gorm.io/gorm"
)

// ───────── cli_password_policy ─────────

func (c *Client) CreatePasswordPolicy(p *db_models.CliPasswordPolicy) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Create(p).Error
}

func (c *Client) GetPasswordPolicyById(id int64) (*db_models.CliPasswordPolicy, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var p db_models.CliPasswordPolicy
	tx := c.Db.First(&p, id)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &p, nil
}

func (c *Client) GetPasswordPolicyByName(name string) (*db_models.CliPasswordPolicy, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var p db_models.CliPasswordPolicy
	tx := c.Db.Where("name = ?", name).First(&p)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &p, nil
}

func (c *Client) ListPasswordPolicies() ([]*db_models.CliPasswordPolicy, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var out []*db_models.CliPasswordPolicy
	return out, c.Db.Find(&out).Error
}

func (c *Client) UpdatePasswordPolicy(p *db_models.CliPasswordPolicy) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Save(p).Error
}

func (c *Client) DeletePasswordPolicyById(id int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Delete(&db_models.CliPasswordPolicy{}, id).Error
}

// ───────── cli_password_history ─────────

func (c *Client) AppendPasswordHistory(h *db_models.CliPasswordHistory) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Create(h).Error
}

func (c *Client) GetRecentPasswordHistory(userID int64, limit int) ([]*db_models.CliPasswordHistory, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var out []*db_models.CliPasswordHistory
	q := c.Db.Where("user_id = ?", userID).Order("changed_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	return out, q.Find(&out).Error
}

// PrunePasswordHistory keeps only the most-recent `keep` rows for user.
func (c *Client) PrunePasswordHistory(userID int64, keep int) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	if keep <= 0 {
		return c.Db.Where("user_id = ?", userID).Delete(&db_models.CliPasswordHistory{}).Error
	}
	// Delete rows older than the newest `keep`. Subquery-based on purpose so
	// we don't fetch-then-delete (race + wasted round trip).
	sub := c.Db.Model(&db_models.CliPasswordHistory{}).
		Select("id").Where("user_id = ?", userID).
		Order("changed_at DESC").Limit(keep)
	return c.Db.Where("user_id = ? AND id NOT IN (?)", userID, sub).
		Delete(&db_models.CliPasswordHistory{}).Error
}

// ───────── cli_group_mgt_permission ─────────

func (c *Client) CreateMgtPermission(p *db_models.CliGroupMgtPermission) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Create(p).Error
}

func (c *Client) ListMgtPermissions(groupID int64) ([]*db_models.CliGroupMgtPermission, error) {
	if c.Db == nil {
		return nil, errors.New("no database connection")
	}
	var out []*db_models.CliGroupMgtPermission
	return out, c.Db.Where("group_id = ?", groupID).Find(&out).Error
}

func (c *Client) DeleteMgtPermissionById(id int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Delete(&db_models.CliGroupMgtPermission{}, id).Error
}

func (c *Client) DeleteAllMgtPermissionByGroupId(groupID int64) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Where("group_id = ?", groupID).Delete(&db_models.CliGroupMgtPermission{}).Error
}
