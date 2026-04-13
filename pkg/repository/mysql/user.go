package mysql

import (
	"errors"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"gorm.io/gorm"
)

func (c *Client) GetAllUser() ([]*db_models.TblAccount, error) {
	var userList []*db_models.TblAccount
	tx := c.Db.Find(&userList)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return userList, nil
}

func (c *Client) GetUserByUserName(username string) (*db_models.TblAccount, error) {
	cond := &db_models.TblAccount{AccountName: username}
	result := &db_models.TblAccount{}
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

func (c *Client) AddUser(user *db_models.TblAccount) error {
	tx := c.Db.Create(user)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (c *Client) UpdateUser(user *db_models.TblAccount) error {
	tx := c.Db.Save(user)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (c *Client) UpdateLoginHistory(username, ipAddress string, timeLogin time.Time) error {
	cond := &db_models.CliLoginHistory{
		UserName:  username,
		IPAddress: ipAddress,
		TimeLogin: timeLogin,
	}
	tx := c.Db.Create(cond)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}
