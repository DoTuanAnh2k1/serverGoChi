package mongodb

import (
	"context"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	colAccounts         = "tbl_account"
	colNe               = "cli_ne"
	colRole             = "cli_role"
	colRoleUserMapping  = "cli_role_user_mapping"
	colUserNeMapping    = "cli_user_ne_mapping"
	colOperationHistory = "cli_operation_history"
	colLoginHistory     = "cli_login_history"
	// colConfigBackup được định nghĩa trong config_backup.go
)

type Client struct {
	Db *mongo.Database
}

var client *Client

func GetInstance() *Client {
	if client == nil {
		client = &Client{}
	}
	return client
}

func (c *Client) Init(cfg config_models.DatabaseConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Client().ApplyURI(cfg.Mongo.URI)
	mc, err := mongo.Connect(ctx, opts)
	if err != nil {
		logger.Logger.WithField("uri", cfg.Mongo.URI).Errorf("mongodb: connect: %v", err)
		return err
	}
	if err = mc.Ping(ctx, readpref.Primary()); err != nil {
		logger.Logger.WithField("uri", cfg.Mongo.URI).Errorf("mongodb: ping: %v", err)
		return err
	}
	c.Db = mc.Database(cfg.Mongo.Database)
	logger.Logger.WithField("db", cfg.Mongo.Database).Info("mongodb: connected")
	return nil
}

func (c *Client) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.Db.Client().Ping(ctx, readpref.Primary())
}

func (c *Client) col(name string) *mongo.Collection {
	return c.Db.Collection(name)
}
