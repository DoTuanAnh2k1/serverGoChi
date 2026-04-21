package mongodb

import (
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (c *Client) GetNeListById(id int64) ([]*db_models.CliNe, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
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
	ctx, cancel := c.opCtx()
	defer cancel()
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
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mNe
	err := c.col(colNe).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMNe(&m), nil
}

// GetNeMonitorById derives monitor info from CliNe — CommandURL is used as the monitor URL.
func (c *Client) GetNeMonitorById(id int64) (*db_models.CliNeMonitor, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mNe
	err := c.col(colNe).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ne := fromMNe(&m)
	return &db_models.CliNeMonitor{
		NeID:      ne.ID,
		NeName:    ne.NeName,
		NeIP:      ne.CommandURL,
		Namespace: ne.Namespace,
	}, nil
}

func (c *Client) GetCLIUserNeMappingByUserId(userId int64) (*db_models.CliUserNeMapping, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mUserNeMapping
	err := c.col(colUserNeMapping).FindOne(ctx, bson.M{"user_id": userId}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMUserNeMapping(&m), nil
}

func (c *Client) DeleteCliNeById(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colNe).DeleteOne(ctx, bson.M{"id": id})
	return err
}

func (c *Client) CreateCliNe(ne *db_models.CliNe) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if ne.ID == 0 {
		id, err := c.nextID(ctx, colNe)
		if err != nil {
			return err
		}
		ne.ID = id
	}
	_, err := c.col(colNe).InsertOne(ctx, toMNe(ne))
	return err
}

func (c *Client) UpdateCliNe(ne *db_models.CliNe) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	filter := bson.M{"id": ne.ID}
	update := bson.M{"$set": toMNe(ne)}
	_, err := c.col(colNe).UpdateOne(ctx, filter, update)
	return err
}

func (c *Client) CreateUserNeMapping(mapping *db_models.CliUserNeMapping) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	doc := bson.M{"user_id": mapping.UserID, "tbl_ne_id": mapping.TblNeID}
	_, err := c.col(colUserNeMapping).InsertOne(ctx, doc)
	return err
}

func (c *Client) DeleteUserNeMapping(mapping *db_models.CliUserNeMapping) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	filter := bson.M{"user_id": mapping.UserID, "tbl_ne_id": mapping.TblNeID}
	_, err := c.col(colUserNeMapping).DeleteOne(ctx, filter)
	return err
}

func (c *Client) GetAllNeOfUserByUserId(userId int64) ([]*db_models.CliUserNeMapping, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
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
