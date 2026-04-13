package mongodb

import (
	"context"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const colConfigBackup = "cli_config_backup"

func (c *Client) SaveConfigBackup(b *db_models.CliConfigBackup) error {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now().UTC()
	}
	// MongoDB has no auto-increment — use nanosecond timestamp as unique int64 ID.
	if b.ID == 0 {
		b.ID = b.CreatedAt.UnixNano()
	}
	_, err := c.col(colConfigBackup).InsertOne(context.Background(), toMConfigBackup(b))
	return err
}

func (c *Client) ListConfigBackups(neName string) ([]*db_models.CliConfigBackup, error) {
	ctx := context.Background()
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
	var m mConfigBackup
	err := c.col(colConfigBackup).FindOne(context.Background(), bson.M{"id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMConfigBackup(&m), nil
}
