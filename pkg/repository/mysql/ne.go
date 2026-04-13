package mysql

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

// GetNeMonitorById derives monitor info from CliNe — CommandURL is used as the monitor URL.
func (c *Client) GetNeMonitorById(id int64) (*db_models.CliNeMonitor, error) {
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
	return &db_models.CliNeMonitor{
		NeID:      ne.ID,
		NeName:    ne.NeName,
		NeIP:      ne.CommandURL,
		Namespace: ne.Namespace,
	}, nil
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

func (c *Client) DeleteCliNeById(id int64) error {
	tx := c.Db.Delete(&db_models.CliNe{}, id)
	if tx == nil {
		return errors.New("no database connection")
	}
	return tx.Error
}

func (c *Client) CreateCliNe(ne *db_models.CliNe) error {
	if c.Db == nil {
		return errors.New("no database connection")
	}
	return c.Db.Create(ne).Error
}

func (c *Client) UpdateCliNe(ne *db_models.CliNe) error {
	tx := c.Db.Save(ne)
	if tx == nil {
		return errors.New("no database connection")
	}
	return tx.Error
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
