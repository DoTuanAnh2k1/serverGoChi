package mongodb

import (
	"context"
	"strings"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"go.mongodb.org/mongo-driver/bson"
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

	// ApplyURI first so any options the user sets in the connection string
	// (readPreference, writeConcern, replicaSet, authSource, tls, ...) win.
	// Code-level Set* calls below only fill in safe defaults when the URI
	// is silent. Write concern intentionally has no default here — for a
	// 2-node replica set `w=majority` deadlocks when one node is down, so
	// operators must choose w=1 (2-node) or w=majority (3+-node) via URI.
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

// Ping uses PrimaryPreferred so a brief primary election doesn't immediately
// flip the health probe to failing.
func (c *Client) Ping() error {
	ctx, cancel := c.opCtx()
	defer cancel()
	return c.Db.Client().Ping(ctx, readpref.PrimaryPreferred())
}

func (c *Client) col(name string) *mongo.Collection {
	return c.Db.Collection(name)
}

// opCtx returns a bounded context for a single DB operation so callers don't
// hang indefinitely during primary election or network partition.
func (c *Client) opCtx() (context.Context, context.CancelFunc) {
	d := c.opTimeout
	if d == 0 {
		d = defaultOpTimeout
	}
	return context.WithTimeout(context.Background(), d)
}

// ensureIndexes creates the unique + performance indexes that MySQL/Postgres
// get from db.sql. Idempotent — Mongo silently skips already-present indexes.
// Also drops obsolete indexes from previous schema versions so upgrades work.
func (c *Client) ensureIndexes(ctx context.Context) error {
	// Obsolete indexes to drop on startup. Keep this list small and use
	// SetName when creating indexes so upgrades can target them by name.
	obsolete := map[string][]string{
		colNe: {"uq_ne_name"}, // replaced by uq_ne_name_namespace
	}
	for coll, names := range obsolete {
		for _, name := range names {
			if _, err := c.col(coll).Indexes().DropOne(ctx, name); err != nil &&
				!strings.Contains(err.Error(), "index not found") &&
				!strings.Contains(err.Error(), "ns not found") {
				return err
			}
		}
	}

	plan := map[string][]mongo.IndexModel{
		colAccounts: {
			{Keys: bson.D{{Key: "account_id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_account_id")},
			{Keys: bson.D{{Key: "account_name", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_account_name")},
		},
		colNe: {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_ne_id")},
			// Allow same ne_name across different namespaces (e.g. per-tenant NEs).
			{Keys: bson.D{{Key: "ne_name", Value: 1}, {Key: "namespace", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_ne_name_namespace")},
			{Keys: bson.D{{Key: "system_type", Value: 1}}, Options: options.Index().SetName("ix_ne_system_type")},
		},
		colGroup: {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_group_id")},
			{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_group_name")},
		},
		colUserNeMapping: {
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "tbl_ne_id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_user_ne")},
			{Keys: bson.D{{Key: "tbl_ne_id", Value: 1}}, Options: options.Index().SetName("ix_user_ne_tbl_ne_id")},
		},
		colUserGroupMapping: {
			{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "group_id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_user_group")},
			{Keys: bson.D{{Key: "group_id", Value: 1}}, Options: options.Index().SetName("ix_user_group_group_id")},
		},
		colGroupNeMapping: {
			{Keys: bson.D{{Key: "group_id", Value: 1}, {Key: "tbl_ne_id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_group_ne")},
			{Keys: bson.D{{Key: "tbl_ne_id", Value: 1}}, Options: options.Index().SetName("ix_group_ne_tbl_ne_id")},
		},
		colOperationHistory: {
			{Keys: bson.D{{Key: "created_date", Value: -1}}, Options: options.Index().SetName("ix_history_created_date")},
			{Keys: bson.D{{Key: "ne_name", Value: 1}}, Options: options.Index().SetName("ix_history_ne_name")},
			{Keys: bson.D{{Key: "account", Value: 1}}, Options: options.Index().SetName("ix_history_account")},
			{Keys: bson.D{{Key: "scope", Value: 1}}, Options: options.Index().SetName("ix_history_scope")},
		},
		colLoginHistory: {
			{Keys: bson.D{{Key: "user_name", Value: 1}}, Options: options.Index().SetName("ix_login_user_name")},
			{Keys: bson.D{{Key: "time_login", Value: -1}}, Options: options.Index().SetName("ix_login_time_login")},
		},
		colConfigBackup: {
			{Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true).SetName("uq_cfgbk_id")},
			{Keys: bson.D{{Key: "ne_name", Value: 1}}, Options: options.Index().SetName("ix_cfgbk_ne_name")},
			{Keys: bson.D{{Key: "created_at", Value: -1}}, Options: options.Index().SetName("ix_cfgbk_created_at")},
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
