package mongodb

import (
	"context"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (c *Client) GetCliRole(cliRole *db_models.CliRole) (*db_models.CliRole, error) {
	filter := bson.M{}
	if cliRole.RoleID != 0 {
		filter["role_id"] = cliRole.RoleID
	}
	if cliRole.NeType != "" {
		filter["ne_type"] = cliRole.NeType
	}
	if cliRole.Scope != "" {
		filter["scope"] = cliRole.Scope
	}
	if cliRole.Permission != "" {
		filter["permission"] = cliRole.Permission
	}

	var m mRole
	err := c.col(colRole).FindOne(context.Background(), filter).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMRole(&m), nil
}

func (c *Client) CreateCliRole(cliRole *db_models.CliRole) error {
	_, err := c.col(colRole).InsertOne(context.Background(), toMRole(cliRole))
	return err
}

func (c *Client) DeleteCliRole(cliRole *db_models.CliRole) error {
	_, err := c.col(colRole).DeleteOne(context.Background(), bson.M{"role_id": cliRole.RoleID})
	return err
}

func (c *Client) GetAllCliRole() ([]*db_models.CliRole, error) {
	ctx := context.Background()
	cur, err := c.col(colRole).Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []*db_models.CliRole
	for cur.Next(ctx) {
		var m mRole
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMRole(&m))
	}
	return results, cur.Err()
}

func (c *Client) GetRolesById(id int64) ([]*db_models.CliRoleUserMapping, error) {
	ctx := context.Background()
	cur, err := c.col(colRoleUserMapping).Find(ctx, bson.M{"user_id": id})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var results []*db_models.CliRoleUserMapping
	for cur.Next(ctx) {
		var m mRoleUserMapping
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMRoleUserMapping(&m))
	}
	return results, cur.Err()
}

func (c *Client) AddRole(role *db_models.CliRoleUserMapping) error {
	filter := bson.M{"user_id": role.UserID, "permission": role.Permission}
	doc := bson.M{"$set": bson.M{"user_id": role.UserID, "permission": role.Permission}}
	opts := options.Update().SetUpsert(true)
	_, err := c.col(colRoleUserMapping).UpdateOne(context.Background(), filter, doc, opts)
	return err
}

func (c *Client) DeleteRole(role *db_models.CliRoleUserMapping) error {
	filter := bson.M{"user_id": role.UserID, "permission": role.Permission}
	_, err := c.col(colRoleUserMapping).DeleteOne(context.Background(), filter)
	return err
}
