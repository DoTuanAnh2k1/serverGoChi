package mongodb

import (
	"context"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (c *Client) GetNeListById(id int64) ([]*db_models.CliNe, error) {
	ctx := context.Background()
	cur, err := c.col(colNe).Find(ctx, bson.M{"id": id})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []*db_models.CliNe
	for cur.Next(ctx) {
		var m mNe
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMNe(&m))
	}
	return results, cur.Err()
}

func (c *Client) GetCliNeListBySystemType(systemType string) ([]*db_models.CliNe, error) {
	ctx := context.Background()
	cur, err := c.col(colNe).Find(ctx, bson.M{"system_type": systemType})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []*db_models.CliNe
	for cur.Next(ctx) {
		var m mNe
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMNe(&m))
	}
	return results, cur.Err()
}

func (c *Client) GetCliNeByNeId(id int64) (*db_models.CliNe, error) {
	var m mNe
	err := c.col(colNe).FindOne(context.Background(), bson.M{"id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, mongo.ErrNoDocuments
	}
	if err != nil {
		return nil, err
	}
	return fromMNe(&m), nil
}

func (c *Client) GetNeMonitorById(id int64) (*db_models.CliNeMonitor, error) {
	var m mNeMonitor
	err := c.col(colNeMonitor).FindOne(context.Background(), bson.M{"ne_id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMNeMonitor(&m), nil
}

func (c *Client) GetCLIUserNeMappingByUserId(userId int64) (*db_models.CliUserNeMapping, error) {
	var m mUserNeMapping
	err := c.col(colUserNeMapping).FindOne(context.Background(), bson.M{"user_id": userId}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, mongo.ErrNoDocuments
	}
	if err != nil {
		return nil, err
	}
	return fromMUserNeMapping(&m), nil
}

func (c *Client) DeleteCliNeById(id int64) error {
	_, err := c.col(colNe).DeleteOne(context.Background(), bson.M{"id": id})
	return err
}

func (c *Client) CreateCliNe(ne *db_models.CliNe) error {
	_, err := c.col(colNe).InsertOne(context.Background(), toMNe(ne))
	return err
}

func (c *Client) UpdateCliNe(ne *db_models.CliNe) error {
	filter := bson.M{"id": ne.ID}
	update := bson.M{"$set": toMNe(ne)}
	_, err := c.col(colNe).UpdateOne(context.Background(), filter, update)
	return err
}

func (c *Client) CreateUserNeMapping(mapping *db_models.CliUserNeMapping) error {
	doc := bson.M{"user_id": mapping.UserID, "tbl_ne_id": mapping.TblNeID}
	_, err := c.col(colUserNeMapping).InsertOne(context.Background(), doc)
	return err
}

func (c *Client) DeleteUserNeMapping(mapping *db_models.CliUserNeMapping) error {
	filter := bson.M{"user_id": mapping.UserID, "tbl_ne_id": mapping.TblNeID}
	_, err := c.col(colUserNeMapping).DeleteOne(context.Background(), filter)
	return err
}

func (c *Client) GetAllNeOfUserByUserId(userId int64) ([]*db_models.CliUserNeMapping, error) {
	ctx := context.Background()
	cur, err := c.col(colUserNeMapping).Find(ctx, bson.M{"user_id": userId})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []*db_models.CliUserNeMapping
	for cur.Next(ctx) {
		var m mUserNeMapping
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMUserNeMapping(&m))
	}
	return results, cur.Err()
}
