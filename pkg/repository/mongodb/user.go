package mongodb

import (
	"context"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (c *Client) GetAllUser() ([]*db_models.TblAccount, error) {
	ctx := context.Background()
	cur, err := c.col(colAccounts).Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []*db_models.TblAccount
	for cur.Next(ctx) {
		var m mAccount
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMAccount(&m))
	}
	return results, cur.Err()
}

func (c *Client) GetUserByUserName(username string) (*db_models.TblAccount, error) {
	var m mAccount
	err := c.col(colAccounts).FindOne(context.Background(), bson.M{"account_name": username}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, mongo.ErrNoDocuments
	}
	if err != nil {
		return nil, err
	}
	return fromMAccount(&m), nil
}

func (c *Client) AddUser(user *db_models.TblAccount) error {
	_, err := c.col(colAccounts).InsertOne(context.Background(), toMAccount(user))
	return err
}

func (c *Client) UpdateUser(user *db_models.TblAccount) error {
	filter := bson.M{"account_id": user.AccountID}
	update := bson.M{"$set": toMAccount(user)}
	_, err := c.col(colAccounts).UpdateOne(context.Background(), filter, update)
	return err
}

func (c *Client) UpdateLoginHistory(username, ipAddress string, timeLogin time.Time) error {
	doc := bson.M{
		"user_name":  username,
		"ip_address": ipAddress,
		"time_login": timeLogin,
	}
	_, err := c.col(colLoginHistory).InsertOne(context.Background(), doc)
	return err
}
