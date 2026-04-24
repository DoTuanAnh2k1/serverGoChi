package mysql

import (
	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"

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

func (c *Client) Init(cfg config_models.DatabaseConfig) error {
	var (
		DbUsername = cfg.Mysql.User
		DbPassword = cfg.Mysql.Password
		DbHost     = cfg.Mysql.Host
		DbPort     = cfg.Mysql.Port
		DbName     = cfg.Mysql.Name
	)
	gormLogger := logger.NewGormLogger()
	gormLogger.LogMode(1)
	dsn := DbUsername + ":" + DbPassword + "@tcp" + "(" + DbHost + ":" + DbPort + ")/" + DbName + "?" + "parseTime=true&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		logger.Logger.WithField("host", DbHost+":"+DbPort).Errorf("mysql: connect: %v", err)
		return err
	}
	logger.Logger.WithField("host", DbHost+":"+DbPort).WithField("db", DbName).Info("mysql: connected")
	c.Db = db

	if err := c.autoMigrate(); err != nil {
		logger.Logger.Errorf("mysql: auto-migrate: %v", err)
		return err
	}

	return nil
}

func (c *Client) autoMigrate() error {
	return c.Db.AutoMigrate(
		&db_models.User{},
		&db_models.NE{},
		&db_models.Command{},
		&db_models.NeAccessGroup{},
		&db_models.NeAccessGroupUser{},
		&db_models.NeAccessGroupNe{},
		&db_models.CmdExecGroup{},
		&db_models.CmdExecGroupUser{},
		&db_models.CmdExecGroupCommand{},
		&db_models.PasswordPolicy{},
		&db_models.PasswordHistory{},
		&db_models.UserAccessList{},
		&db_models.OperationHistory{},
		&db_models.LoginHistory{},
		&db_models.ConfigBackup{},
	)
}

func (c *Client) Ping() error {
	sql, err := c.Db.DB()
	if err != nil {
		return err
	}
	return sql.Ping()
}
