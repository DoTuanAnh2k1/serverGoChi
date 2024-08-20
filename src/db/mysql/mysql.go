package mysql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"serverGoChi/models/config_models"
	"serverGoChi/src/log"
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
	//var (
	//	DbUsername = cfg.Db.Mysql.User
	//	DbPassword = cfg.Db.Mysql.Password
	//	DbHost     = cfg.Db.Mysql.Host
	//	DbPort     = cfg.Db.Mysql.Port
	//	DbName     = cfg.Db.Mysql.Name
	//)
	var (
		DbUsername = "ems"
		DbPassword = "Ems@2021"
		DbHost     = "127.0.0.1"
		DbPort     = cfg.Db.Mysql.Port
		DbName     = "ems"
	)
	dsn := DbUsername + ":" + DbPassword + "@tcp" + "(" + DbHost + ":" + DbPort + ")/" + DbName + "?" + "parseTime=true&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Logger.Debugf("Error connecting to database : error=%v", err)
		return err
	}
	log.Logger.Info("Connect to database: ", dsn)
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
