package mongodb

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const colConfigBackup = "cli_config_backup"

func (c *Client) SaveConfigBackup(b *db_models.CliConfigBackup) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now().UTC()
	}
	if b.ID == 0 {
		id, err := c.nextID(ctx, colConfigBackup)
		if err != nil {
			return err
		}
		b.ID = id
	}
	_, err := c.col(colConfigBackup).InsertOne(ctx, toMConfigBackup(b))
	return err
}

func (c *Client) ListConfigBackups(neName string) ([]*db_models.CliConfigBackup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	filter := bson.M{}
	if neName != "" {
		filter["ne_name"] = neName
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cur, err := c.col(colConfigBackup).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []*db_models.CliConfigBackup
	for cur.Next(ctx) {
		var m mConfigBackup
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMConfigBackup(&m))
	}
	return results, cur.Err()
}

func (c *Client) GetConfigBackupById(id int64) (*db_models.CliConfigBackup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mConfigBackup
	err := c.col(colConfigBackup).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMConfigBackup(&m), nil
}
