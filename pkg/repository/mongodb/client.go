// Package mongodb implements DatabaseStore on the native Mongo driver.
// Collection names mirror db.sql table names. Sequential int64 ids are
// managed via the counters collection (see counters.go) because Mongo has
// no native AUTO_INCREMENT.
package mongodb

import (
	"context"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// v2 collection names — 1:1 with db.sql table names.
const (
	colUser                = "user"
	colNe                  = "ne"
	colCommand             = "command"
	colNeAccessGroup       = "ne_access_group"
	colNeAccessGroupUser   = "ne_access_group_user"
	colNeAccessGroupNe     = "ne_access_group_ne"
	colCmdExecGroup        = "cmd_exec_group"
	colCmdExecGroupUser    = "cmd_exec_group_user"
	colCmdExecGroupCommand = "cmd_exec_group_command"
	colPasswordPolicy      = "password_policy"
	colPasswordHistory     = "password_history"
	colUserAccessList      = "user_access_list"
	colOperationHistory    = "operation_history"
	colLoginHistory        = "login_history"
	colConfigBackup        = "config_backup"

	defaultOpTimeout = 10 * time.Second
)

type Client struct {
	Db        *mongo.Database
	opTimeout time.Duration
}

var client *Client

func GetInstance() *Client {
	if client == nil {
		client = &Client{opTimeout: defaultOpTimeout}
	}
	return client
}

func (c *Client) Init(cfg config_models.DatabaseConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Client().ApplyURI(cfg.Mongo.URI)
	if opts.ReadPreference == nil {
		opts.SetReadPreference(readpref.PrimaryPreferred())
	}
	if opts.RetryWrites == nil {
		opts.SetRetryWrites(true)
	}
	if opts.RetryReads == nil {
		opts.SetRetryReads(true)
	}
	if opts.ServerSelectionTimeout == nil {
		opts.SetServerSelectionTimeout(15 * time.Second)
	}

	mc, err := mongo.Connect(ctx, opts)
	if err != nil {
		logger.Logger.WithField("uri", cfg.Mongo.URI).Errorf("mongodb: connect: %v", err)
		return err
	}
	if err = mc.Ping(ctx, readpref.PrimaryPreferred()); err != nil {
		logger.Logger.WithField("uri", cfg.Mongo.URI).Errorf("mongodb: ping: %v", err)
		return err
	}
	c.Db = mc.Database(cfg.Mongo.Database)
	if c.opTimeout == 0 {
		c.opTimeout = defaultOpTimeout
	}
	logger.Logger.WithField("db", cfg.Mongo.Database).Info("mongodb: connected")

	if err := c.ensureIndexes(ctx); err != nil {
		logger.Logger.Errorf("mongodb: ensure indexes: %v", err)
		return err
	}
	return nil
}

// Ping uses PrimaryPreferred so a brief primary election doesn't flip the probe.
func (c *Client) Ping() error {
	ctx, cancel := c.opCtx()
	defer cancel()
	return c.Db.Client().Ping(ctx, readpref.PrimaryPreferred())
}

func (c *Client) col(name string) *mongo.Collection {
	return c.Db.Collection(name)
}

func (c *Client) opCtx() (context.Context, context.CancelFunc) {
	d := c.opTimeout
	if d == 0 {
		d = defaultOpTimeout
	}
	return context.WithTimeout(context.Background(), d)
}

// ensureIndexes mirrors the UNIQUE/INDEX declarations in db.sql. Idempotent.
func (c *Client) ensureIndexes(ctx context.Context) error {
	plan := map[string][]mongo.IndexModel{
		colUser: {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_user_id")},
			{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_user_username")},
		},
		colNe: {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_ne_id")},
			{Keys: bson.D{{Key: "namespace", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_ne_namespace")},
			{Keys: bson.D{{Key: "ne_type", Value: 1}}, Options: options.Index().SetName("ix_ne_type")},
		},
		colCommand: {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_command_id")},
			{Keys: bson.D{{Key: "ne_id", Value: 1}, {Key: "service", Value: 1}, {Key: "cmd_text", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_command_sig")},
			{Keys: bson.D{{Key: "ne_id", Value: 1}}, Options: options.Index().SetName("ix_command_ne")},
		},
		colNeAccessGroup: {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_nag_id")},
			{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_nag_name")},
		},
		colNeAccessGroupUser: {
			{Keys: bson.D{{Key: "group_id", Value: 1}, {Key: "user_id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_nag_user")},
			{Keys: bson.D{{Key: "user_id", Value: 1}}, Options: options.Index().SetName("ix_nag_user_uid")},
		},
		colNeAccessGroupNe: {
			{Keys: bson.D{{Key: "group_id", Value: 1}, {Key: "ne_id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_nag_ne")},
			{Keys: bson.D{{Key: "ne_id", Value: 1}}, Options: options.Index().SetName("ix_nag_ne_nid")},
		},
		colCmdExecGroup: {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_ceg_id")},
			{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_ceg_name")},
		},
		colCmdExecGroupUser: {
			{Keys: bson.D{{Key: "group_id", Value: 1}, {Key: "user_id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_ceg_user")},
			{Keys: bson.D{{Key: "user_id", Value: 1}}, Options: options.Index().SetName("ix_ceg_user_uid")},
		},
		colCmdExecGroupCommand: {
			{Keys: bson.D{{Key: "group_id", Value: 1}, {Key: "command_id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_ceg_cmd")},
			{Keys: bson.D{{Key: "command_id", Value: 1}}, Options: options.Index().SetName("ix_ceg_cmd_cid")},
		},
		colPasswordPolicy: {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_pwpol_id")},
		},
		colPasswordHistory: {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_pwh_id")},
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "changed_at", Value: -1}}, Options: options.Index().SetName("ix_pwh_user_changed")},
		},
		colUserAccessList: {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_acl_id")},
			{Keys: bson.D{{Key: "list_type", Value: 1}, {Key: "match_type", Value: 1}, {Key: "pattern", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_acl_sig")},
		},
		colOperationHistory: {
			{Keys: bson.D{{Key: "created_date", Value: -1}}, Options: options.Index().SetName("ix_history_created")},
			{Keys: bson.D{{Key: "account", Value: 1}}, Options: options.Index().SetName("ix_history_account")},
			{Keys: bson.D{{Key: "ne_namespace", Value: 1}}, Options: options.Index().SetName("ix_history_ne")},
			{Keys: bson.D{{Key: "scope", Value: 1}}, Options: options.Index().SetName("ix_history_scope")},
		},
		colLoginHistory: {
			{Keys: bson.D{{Key: "username", Value: 1}}, Options: options.Index().SetName("ix_login_username")},
			{Keys: bson.D{{Key: "time_login", Value: -1}}, Options: options.Index().SetName("ix_login_time")},
		},
		colConfigBackup: {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_cfgbk_id")},
			{Keys: bson.D{{Key: "ne_name", Value: 1}}, Options: options.Index().SetName("ix_cfgbk_ne_name")},
			{Keys: bson.D{{Key: "created_at", Value: -1}}, Options: options.Index().SetName("ix_cfgbk_created")},
		},
		colCounters: {
			{Keys: bson.D{{Key: "_id", Value: 1}}, Options: options.Index().SetName("ix_counters_id")},
		},
	}
	for coll, models := range plan {
		if _, err := c.col(coll).Indexes().CreateMany(ctx, models); err != nil {
			return err
		}
	}
	return nil
}
