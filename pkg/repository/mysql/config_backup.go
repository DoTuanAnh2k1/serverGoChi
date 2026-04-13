package mysql

import (
	"errors"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"gorm.io/gorm"
)

func (c *Client) SaveConfigBackup(b *db_models.CliConfigBackup) error {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now().UTC()
	}
	tx := c.Db.Create(b)
	if tx == nil {
		return errors.New("no database connection")
	}
	return tx.Error
}

func (c *Client) ListConfigBackups(neName string) ([]*db_models.CliConfigBackup, error) {
	var list []*db_models.CliConfigBackup
	q := c.Db.Order("created_at DESC")
	if neName != "" {
		q = q.Where("ne_name = ?", neName)
	}
	tx := q.Find(&list)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	return list, tx.Error
}

func (c *Client) GetConfigBackupById(id int64) (*db_models.CliConfigBackup, error) {
	result := &db_models.CliConfigBackup{}
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
