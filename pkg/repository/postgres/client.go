package postgres

import (
	"fmt"

	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Client struct {
	Db *gorm.DB
}

var client *Client

func GetInstance() *Client {
	if client == nil {
		client = &Client{}
	}
	return client
}

func (c *Client) Init(cfg config_models.DatabaseConfig) error {
	p := cfg.Postgres
	sslMode := p.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		p.Host, p.Port, p.User, p.Password, p.Name, sslMode,
	)

	gormLogger := logger.NewGormLogger()
	gormLogger.LogMode(1)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		logger.Logger.WithField("host", p.Host+":"+p.Port).Errorf("postgres: connect: %v", err)
		return err
	}
	logger.Logger.WithField("host", p.Host+":"+p.Port).WithField("db", p.Name).Info("postgres: connected")
	c.Db = db

	if err := c.autoMigrate(); err != nil {
		logger.Logger.Errorf("postgres: auto-migrate: %v", err)
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
