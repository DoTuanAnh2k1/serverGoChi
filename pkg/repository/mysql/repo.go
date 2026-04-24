// Package mysql implements DatabaseStore on GORM + MySQL/MariaDB. Postgres
// driver is character-for-character identical except for the package
// declaration — see pkg/repository/postgres/repo.go.
package mysql

import (
	"errors"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"gorm.io/gorm"
)

func nilIfNotFound(tx *gorm.DB) error {
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	return tx.Error
}

// ── User ────────────────────────────────────────────────────────────────

func (c *Client) CreateUser(u *db_models.User) error { return c.Db.Create(u).Error }

func (c *Client) GetUserByID(id int64) (*db_models.User, error) {
	var u db_models.User
	tx := c.Db.First(&u, id)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &u, nil
}

func (c *Client) GetUserByUsername(username string) (*db_models.User, error) {
	var u db_models.User
	tx := c.Db.Where("username = ?", username).First(&u)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &u, nil
}

func (c *Client) ListUsers() ([]*db_models.User, error) {
	var out []*db_models.User
	return out, c.Db.Order("id").Find(&out).Error
}

func (c *Client) UpdateUser(u *db_models.User) error { return c.Db.Save(u).Error }

func (c *Client) DeleteUserByID(id int64) error {
	return c.Db.Delete(&db_models.User{}, id).Error
}

// ── NE ──────────────────────────────────────────────────────────────────

func (c *Client) CreateNE(n *db_models.NE) error { return c.Db.Create(n).Error }

func (c *Client) GetNEByID(id int64) (*db_models.NE, error) {
	var n db_models.NE
	tx := c.Db.First(&n, id)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &n, nil
}

func (c *Client) GetNEByNamespace(ns string) (*db_models.NE, error) {
	var n db_models.NE
	tx := c.Db.Where("namespace = ?", ns).First(&n)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &n, nil
}

func (c *Client) ListNEs() ([]*db_models.NE, error) {
	var out []*db_models.NE
	return out, c.Db.Order("id").Find(&out).Error
}

func (c *Client) UpdateNE(n *db_models.NE) error { return c.Db.Save(n).Error }

func (c *Client) DeleteNEByID(id int64) error {
	return c.Db.Delete(&db_models.NE{}, id).Error
}

// ── Command ─────────────────────────────────────────────────────────────

func (c *Client) CreateCommand(cmd *db_models.Command) error { return c.Db.Create(cmd).Error }

func (c *Client) GetCommandByID(id int64) (*db_models.Command, error) {
	var cmd db_models.Command
	tx := c.Db.First(&cmd, id)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &cmd, nil
}

func (c *Client) GetCommandByTriple(neID int64, service, cmdText string) (*db_models.Command, error) {
	var cmd db_models.Command
	tx := c.Db.Where("ne_id = ? AND service = ? AND cmd_text = ?", neID, service, cmdText).First(&cmd)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &cmd, nil
}

func (c *Client) ListCommands(neID int64, service string) ([]*db_models.Command, error) {
	q := c.Db.Model(&db_models.Command{})
	if neID > 0 {
		q = q.Where("ne_id = ?", neID)
	}
	if service != "" {
		q = q.Where("service = ?", service)
	}
	var out []*db_models.Command
	return out, q.Order("ne_id, service, cmd_text").Find(&out).Error
}

func (c *Client) UpdateCommand(cmd *db_models.Command) error { return c.Db.Save(cmd).Error }

func (c *Client) DeleteCommandByID(id int64) error {
	return c.Db.Delete(&db_models.Command{}, id).Error
}

// ── NE Access Group ─────────────────────────────────────────────────────

func (c *Client) CreateNeAccessGroup(g *db_models.NeAccessGroup) error { return c.Db.Create(g).Error }

func (c *Client) GetNeAccessGroupByID(id int64) (*db_models.NeAccessGroup, error) {
	var g db_models.NeAccessGroup
	tx := c.Db.First(&g, id)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &g, nil
}

func (c *Client) GetNeAccessGroupByName(name string) (*db_models.NeAccessGroup, error) {
	var g db_models.NeAccessGroup
	tx := c.Db.Where("name = ?", name).First(&g)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &g, nil
}

func (c *Client) ListNeAccessGroups() ([]*db_models.NeAccessGroup, error) {
	var out []*db_models.NeAccessGroup
	return out, c.Db.Order("id").Find(&out).Error
}

func (c *Client) UpdateNeAccessGroup(g *db_models.NeAccessGroup) error {
	return c.Db.Save(g).Error
}

func (c *Client) DeleteNeAccessGroupByID(id int64) error {
	return c.Db.Delete(&db_models.NeAccessGroup{}, id).Error
}

func (c *Client) AddUserToNeAccessGroup(groupID, userID int64) error {
	return c.Db.Create(&db_models.NeAccessGroupUser{GroupID: groupID, UserID: userID}).Error
}

func (c *Client) RemoveUserFromNeAccessGroup(groupID, userID int64) error {
	return c.Db.Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&db_models.NeAccessGroupUser{}).Error
}

func (c *Client) ListUsersInNeAccessGroup(groupID int64) ([]int64, error) {
	var ids []int64
	tx := c.Db.Model(&db_models.NeAccessGroupUser{}).
		Where("group_id = ?", groupID).Pluck("user_id", &ids)
	return ids, tx.Error
}

func (c *Client) ListNeAccessGroupsOfUser(userID int64) ([]int64, error) {
	var ids []int64
	tx := c.Db.Model(&db_models.NeAccessGroupUser{}).
		Where("user_id = ?", userID).Pluck("group_id", &ids)
	return ids, tx.Error
}

func (c *Client) AddNeToNeAccessGroup(groupID, neID int64) error {
	return c.Db.Create(&db_models.NeAccessGroupNe{GroupID: groupID, NeID: neID}).Error
}

func (c *Client) RemoveNeFromNeAccessGroup(groupID, neID int64) error {
	return c.Db.Where("group_id = ? AND ne_id = ?", groupID, neID).
		Delete(&db_models.NeAccessGroupNe{}).Error
}

func (c *Client) ListNEsInNeAccessGroup(groupID int64) ([]int64, error) {
	var ids []int64
	tx := c.Db.Model(&db_models.NeAccessGroupNe{}).
		Where("group_id = ?", groupID).Pluck("ne_id", &ids)
	return ids, tx.Error
}

func (c *Client) ListNeAccessGroupsOfNE(neID int64) ([]int64, error) {
	var ids []int64
	tx := c.Db.Model(&db_models.NeAccessGroupNe{}).
		Where("ne_id = ?", neID).Pluck("group_id", &ids)
	return ids, tx.Error
}

// ── Cmd Exec Group ──────────────────────────────────────────────────────

func (c *Client) CreateCmdExecGroup(g *db_models.CmdExecGroup) error { return c.Db.Create(g).Error }

func (c *Client) GetCmdExecGroupByID(id int64) (*db_models.CmdExecGroup, error) {
	var g db_models.CmdExecGroup
	tx := c.Db.First(&g, id)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &g, nil
}

func (c *Client) GetCmdExecGroupByName(name string) (*db_models.CmdExecGroup, error) {
	var g db_models.CmdExecGroup
	tx := c.Db.Where("name = ?", name).First(&g)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &g, nil
}

func (c *Client) ListCmdExecGroups() ([]*db_models.CmdExecGroup, error) {
	var out []*db_models.CmdExecGroup
	return out, c.Db.Order("id").Find(&out).Error
}

func (c *Client) UpdateCmdExecGroup(g *db_models.CmdExecGroup) error { return c.Db.Save(g).Error }

func (c *Client) DeleteCmdExecGroupByID(id int64) error {
	return c.Db.Delete(&db_models.CmdExecGroup{}, id).Error
}

func (c *Client) AddUserToCmdExecGroup(groupID, userID int64) error {
	return c.Db.Create(&db_models.CmdExecGroupUser{GroupID: groupID, UserID: userID}).Error
}

func (c *Client) RemoveUserFromCmdExecGroup(groupID, userID int64) error {
	return c.Db.Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&db_models.CmdExecGroupUser{}).Error
}

func (c *Client) ListUsersInCmdExecGroup(groupID int64) ([]int64, error) {
	var ids []int64
	tx := c.Db.Model(&db_models.CmdExecGroupUser{}).
		Where("group_id = ?", groupID).Pluck("user_id", &ids)
	return ids, tx.Error
}

func (c *Client) ListCmdExecGroupsOfUser(userID int64) ([]int64, error) {
	var ids []int64
	tx := c.Db.Model(&db_models.CmdExecGroupUser{}).
		Where("user_id = ?", userID).Pluck("group_id", &ids)
	return ids, tx.Error
}

func (c *Client) AddCommandToCmdExecGroup(groupID, commandID int64) error {
	return c.Db.Create(&db_models.CmdExecGroupCommand{GroupID: groupID, CommandID: commandID}).Error
}

func (c *Client) RemoveCommandFromCmdExecGroup(groupID, commandID int64) error {
	return c.Db.Where("group_id = ? AND command_id = ?", groupID, commandID).
		Delete(&db_models.CmdExecGroupCommand{}).Error
}

func (c *Client) ListCommandsInCmdExecGroup(groupID int64) ([]int64, error) {
	var ids []int64
	tx := c.Db.Model(&db_models.CmdExecGroupCommand{}).
		Where("group_id = ?", groupID).Pluck("command_id", &ids)
	return ids, tx.Error
}

func (c *Client) ListCmdExecGroupsOfCommand(commandID int64) ([]int64, error) {
	var ids []int64
	tx := c.Db.Model(&db_models.CmdExecGroupCommand{}).
		Where("command_id = ?", commandID).Pluck("group_id", &ids)
	return ids, tx.Error
}

// ── Password Policy / History ───────────────────────────────────────────

func (c *Client) GetPasswordPolicy() (*db_models.PasswordPolicy, error) {
	var p db_models.PasswordPolicy
	tx := c.Db.First(&p, 1)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &p, nil
}

// UpsertPasswordPolicy treats the policy as a singleton at id=1.
func (c *Client) UpsertPasswordPolicy(p *db_models.PasswordPolicy) error {
	p.ID = 1
	var existing db_models.PasswordPolicy
	if err := c.Db.First(&existing, 1).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Db.Create(p).Error
		}
		return err
	}
	return c.Db.Save(p).Error
}

func (c *Client) AppendPasswordHistory(h *db_models.PasswordHistory) error {
	return c.Db.Create(h).Error
}

func (c *Client) GetRecentPasswordHistory(userID int64, limit int) ([]*db_models.PasswordHistory, error) {
	var out []*db_models.PasswordHistory
	q := c.Db.Where("user_id = ?", userID).Order("changed_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	return out, q.Find(&out).Error
}

func (c *Client) PrunePasswordHistory(userID int64, keep int) error {
	if keep <= 0 {
		return c.Db.Where("user_id = ?", userID).Delete(&db_models.PasswordHistory{}).Error
	}
	sub := c.Db.Model(&db_models.PasswordHistory{}).
		Select("id").Where("user_id = ?", userID).
		Order("changed_at DESC").Limit(keep)
	return c.Db.Where("user_id = ? AND id NOT IN (?)", userID, sub).
		Delete(&db_models.PasswordHistory{}).Error
}

// ── User Access List ────────────────────────────────────────────────────

func (c *Client) CreateAccessListEntry(e *db_models.UserAccessList) error {
	return c.Db.Create(e).Error
}

func (c *Client) ListAccessListEntries(listType string) ([]*db_models.UserAccessList, error) {
	var out []*db_models.UserAccessList
	q := c.Db.Model(&db_models.UserAccessList{})
	if listType != "" {
		q = q.Where("list_type = ?", listType)
	}
	return out, q.Order("id").Find(&out).Error
}

func (c *Client) DeleteAccessListEntryByID(id int64) error {
	return c.Db.Delete(&db_models.UserAccessList{}, id).Error
}

// ── History ────────────────────────────────────────────────────────────

func (c *Client) SaveOperationHistory(h db_models.OperationHistory) error {
	return c.Db.Create(&h).Error
}

func (c *Client) GetRecentHistory(limit int) ([]db_models.OperationHistory, error) {
	var out []db_models.OperationHistory
	q := c.Db.Order("created_date DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	return out, q.Find(&out).Error
}

func (c *Client) GetRecentHistoryFiltered(limit int, scope, neNamespace, account string) ([]db_models.OperationHistory, error) {
	q := c.Db.Model(&db_models.OperationHistory{})
	if scope != "" {
		q = q.Where("scope = ?", scope)
	}
	if neNamespace != "" {
		q = q.Where("ne_namespace = ?", neNamespace)
	}
	if account != "" {
		q = q.Where("account = ?", account)
	}
	q = q.Order("created_date DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	var out []db_models.OperationHistory
	return out, q.Find(&out).Error
}

func (c *Client) GetDailyOperationHistory(date time.Time) ([]db_models.OperationHistory, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)
	var out []db_models.OperationHistory
	return out, c.Db.Where("created_date >= ? AND created_date < ?", start, end).
		Order("ne_namespace, created_date").Find(&out).Error
}

func (c *Client) DeleteHistoryBefore(cutoff time.Time) (int64, error) {
	tx := c.Db.Where("created_date < ?", cutoff).Delete(&db_models.OperationHistory{})
	return tx.RowsAffected, tx.Error
}

func (c *Client) UpdateLoginHistory(username, ip string, t time.Time) error {
	return c.Db.Create(&db_models.LoginHistory{
		Username: username, IPAddress: ip, TimeLogin: t,
	}).Error
}

// ── Config Backup ───────────────────────────────────────────────────────

func (c *Client) SaveConfigBackup(b *db_models.ConfigBackup) error { return c.Db.Create(b).Error }

func (c *Client) ListConfigBackups(neName string) ([]*db_models.ConfigBackup, error) {
	q := c.Db.Model(&db_models.ConfigBackup{})
	if neName != "" {
		q = q.Where("ne_name = ?", neName)
	}
	var out []*db_models.ConfigBackup
	return out, q.Order("created_at DESC").Find(&out).Error
}

func (c *Client) GetConfigBackupByID(id int64) (*db_models.ConfigBackup, error) {
	var b db_models.ConfigBackup
	tx := c.Db.First(&b, id)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, tx.Error
	}
	return &b, nil
}

// quiet unused import warnings for files that share this package
var _ = nilIfNotFound
