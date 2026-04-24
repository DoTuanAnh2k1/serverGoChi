package mongodb

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (c *Client) GetAllUser() ([]*db_models.TblAccount, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
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
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mAccount
	err := c.col(colAccounts).FindOne(ctx, bson.M{"account_name": username}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMAccount(&m), nil
}

func (c *Client) AddUser(user *db_models.TblAccount) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if user.AccountID == 0 {
		id, err := c.nextID(ctx, colAccounts)
		if err != nil {
			return err
		}
		user.AccountID = id
	}
	_, err := c.col(colAccounts).InsertOne(ctx, toMAccount(user))
	return err
}

func (c *Client) UpdateUser(user *db_models.TblAccount) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	filter := bson.M{"account_id": user.AccountID}
	update := bson.M{"$set": toMAccount(user)}
	_, err := c.col(colAccounts).UpdateOne(ctx, filter, update)
	return err
}

// DeleteUserById hard-deletes the account document. Callers fan out the
// cascade (user-group / user-ne mappings, password history) first; the
// repo only touches the tbl_account collection.
func (c *Client) DeleteUserById(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colAccounts).DeleteOne(ctx, bson.M{"account_id": id})
	if err == mongo.ErrNoDocuments {
		return nil
	}
	return err
}

// DeleteAllUserNeMappingByUserId is the user-side companion to the NE-side
// existing method, used by PurgeUser.
func (c *Client) DeleteAllUserNeMappingByUserId(userId int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colUserNeMapping).DeleteMany(ctx, bson.M{"user_id": userId})
	return err
}

func (c *Client) UpdateLoginHistory(username, ipAddress string, timeLogin time.Time) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	doc := bson.M{
		"user_name":  username,
		"ip_address": ipAddress,
		"time_login": timeLogin,
	}
	_, err := c.col(colLoginHistory).InsertOne(ctx, doc)
	return err
}
