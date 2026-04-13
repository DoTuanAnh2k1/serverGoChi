package mongodb

import (
	"context"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const colNeConfig = "cli_ne_config"

func (c *Client) CreateCliNeConfig(cfg *db_models.CliNeConfig) error {
	_, err := c.col(colNeConfig).InsertOne(context.Background(), toMNeConfig(cfg))
	return err
}

func (c *Client) GetCliNeConfigByNeId(neId int64) ([]*db_models.CliNeConfig, error) {
	ctx := context.Background()
	cur, err := c.col(colNeConfig).Find(ctx, bson.M{"ne_id": neId})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []*db_models.CliNeConfig
	for cur.Next(ctx) {
		var m mNeConfig
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMNeConfig(&m))
	}
	return results, cur.Err()
}

func (c *Client) GetCliNeConfigById(id int64) (*db_models.CliNeConfig, error) {
	var m mNeConfig
	err := c.col(colNeConfig).FindOne(context.Background(), bson.M{"id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMNeConfig(&m), nil
}

func (c *Client) UpdateCliNeConfig(cfg *db_models.CliNeConfig) error {
	filter := bson.M{"id": cfg.ID}
	update := bson.M{"$set": toMNeConfig(cfg)}
	_, err := c.col(colNeConfig).UpdateOne(context.Background(), filter, update)
	return err
}

func (c *Client) DeleteCliNeConfigById(id int64) error {
	_, err := c.col(colNeConfig).DeleteOne(context.Background(), bson.M{"id": id})
	return err
}

func (c *Client) DeleteCliNeConfigByNeId(neId int64) error {
	_, err := c.col(colNeConfig).DeleteMany(context.Background(), bson.M{"ne_id": neId})
	return err
}

// cascade helpers

func (c *Client) DeleteAllUserNeMappingByNeId(neId int64) error {
	_, err := c.col(colUserNeMapping).DeleteMany(context.Background(), bson.M{"tbl_ne_id": neId})
	return err
}

func (c *Client) DeleteNeMonitorByNeId(neId int64) error {
	_, err := c.col(colNeMonitor).DeleteMany(context.Background(), bson.M{"ne_id": neId})
	return err
}

func (c *Client) DeleteCliNeSlaveByNeId(neId int64) error {
	_, err := c.col("cli_ne_slave").DeleteMany(context.Background(), bson.M{"ne_id": neId})
	return err
}
