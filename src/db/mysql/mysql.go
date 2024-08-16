package mysql

import (
	"errors"
	"serverGoChi/models/config_models"
	"serverGoChi/models/db_models"
	"serverGoChi/src/log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Client struct {
	Db *gorm.DB
}

var (
	client *Client
)

func GetInstance() *Client {
	if client == nil {
		client = &Client{}
	}
	return client
}

func (c *Client) Init(cfg *config_models.Config) error {
	var err error
	var (
		DbUsername = cfg.Db.Mysql.User
		DbPassword = cfg.Db.Mysql.Password
		DbHost     = cfg.Db.Mysql.Host
		DbPort     = cfg.Db.Mysql.Port
		DbName     = cfg.Db.Mysql.Name
	)
	dsn := DbUsername + ":" + DbPassword + "@tcp" + "(" + DbHost + ":" + DbPort + ")/" + DbName + "?" + "parseTime=true&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Logger.Debugf("Error connecting to database : error=%v", err)
		return err
	}
	c.Db = db
	return nil
}

func (c *Client) Ping() error {
	sql, err := c.Db.DB()
	if err != nil {
		return err
	}
	return sql.Ping()
}

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
		return nil, tx.Error
	}
	return result, nil
}

func (c *Client) AddUser(user db_models.TblAccount) error {
	cond := &user
	tx := c.Db.Create(cond)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (c *Client) UpdateUser(user db_models.TblAccount) error {
	cond := &user
	tx := c.Db.Updates(cond)
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

func (c *Client) SaveHistoryCommand(history db_models.CliOperationHistory) error {
	cond := &history
	tx := c.Db.Create(cond)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
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

func (c *Client) GetNeListById(id int64) ([]*db_models.CliNe, error) {
	cond := &db_models.CliNe{ID: id}
	var userList []*db_models.CliNe
	tx := c.Db.Find(userList, cond)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return userList, nil
}

func (c *Client) GetRolesById(id int64) ([]*db_models.CliRoleUserMapping, error) {
	cond := &db_models.CliRoleUserMapping{
		UserID: id,
	}
	var roleList []*db_models.CliRoleUserMapping
	tx := c.Db.Find(roleList, cond)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return roleList, nil
}

func (c *Client) GetCliRole(cliRole db_models.CliRole) (*db_models.CliRole, error) {
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

func (c *Client) CreateCliRole(cliRole db_models.CliRole) error {
	cond := &cliRole
	tx := c.Db.Create(cond)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (c *Client) DeleteCliRole(cliRole db_models.CliRole) error {
	cond := &cliRole
	tx := c.Db.Delete(cond)
	if tx == nil {
		return errors.New("no database connection")
	}
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

func (c *Client) GetAllCliRole() ([]*db_models.CliRole, error) {
	var cliRoleList []*db_models.CliRole
	tx := c.Db.Find(cliRoleList)
	if tx == nil {
		return nil, errors.New("no database connection")
	}
	if tx.Error != nil {
		return nil, tx.Error
	}
	return cliRoleList, nil
}
