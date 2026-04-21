package mongodb

import (
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	colGroup            = "cli_group"
	colUserGroupMapping = "cli_user_group_mapping"
	colGroupNeMapping   = "cli_group_ne_mapping"
)

// ── cli_group CRUD ───────────────────────────────────────────────────────────

type mGroup struct {
	ID          int64  `bson:"id"`
	Name        string `bson:"name"`
	Description string `bson:"description"`
}

func toMGroup(g *db_models.CliGroup) *mGroup {
	return &mGroup{ID: g.ID, Name: g.Name, Description: g.Description}
}

func fromMGroup(m *mGroup) *db_models.CliGroup {
	return &db_models.CliGroup{ID: m.ID, Name: m.Name, Description: m.Description}
}

func (c *Client) CreateGroup(g *db_models.CliGroup) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if g.ID == 0 {
		id, err := c.nextID(ctx, colGroup)
		if err != nil {
			return err
		}
		g.ID = id
	}
	_, err := c.col(colGroup).InsertOne(ctx, toMGroup(g))
	return err
}

func (c *Client) GetGroupById(id int64) (*db_models.CliGroup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mGroup
	err := c.col(colGroup).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMGroup(&m), nil
}

func (c *Client) GetGroupByName(name string) (*db_models.CliGroup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mGroup
	err := c.col(colGroup).FindOne(ctx, bson.M{"name": name}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMGroup(&m), nil
}

func (c *Client) GetAllGroups() ([]*db_models.CliGroup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colGroup).Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var results []*db_models.CliGroup
	for cur.Next(ctx) {
		var m mGroup
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMGroup(&m))
	}
	return results, cur.Err()
}

func (c *Client) UpdateGroup(g *db_models.CliGroup) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colGroup).UpdateOne(ctx, bson.M{"id": g.ID}, bson.M{"$set": toMGroup(g)})
	return err
}

func (c *Client) DeleteGroupById(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colGroup).DeleteOne(ctx, bson.M{"id": id})
	return err
}

// ── cli_user_group_mapping ───────────────────────────────────────────────────

type mUserGroupMapping struct {
	UserID  int64 `bson:"user_id"`
	GroupID int64 `bson:"group_id"`
}

func fromMUserGroupMapping(m *mUserGroupMapping) *db_models.CliUserGroupMapping {
	return &db_models.CliUserGroupMapping{UserID: m.UserID, GroupID: m.GroupID}
}

func (c *Client) CreateUserGroupMapping(m *db_models.CliUserGroupMapping) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colUserGroupMapping).InsertOne(ctx, bson.M{"user_id": m.UserID, "group_id": m.GroupID})
	return err
}

func (c *Client) DeleteUserGroupMapping(m *db_models.CliUserGroupMapping) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colUserGroupMapping).DeleteOne(ctx, bson.M{"user_id": m.UserID, "group_id": m.GroupID})
	return err
}

func (c *Client) GetAllGroupsOfUser(userId int64) ([]*db_models.CliUserGroupMapping, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colUserGroupMapping).Find(ctx, bson.M{"user_id": userId})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var results []*db_models.CliUserGroupMapping
	for cur.Next(ctx) {
		var m mUserGroupMapping
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMUserGroupMapping(&m))
	}
	return results, cur.Err()
}

func (c *Client) GetAllUsersOfGroup(groupId int64) ([]*db_models.CliUserGroupMapping, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colUserGroupMapping).Find(ctx, bson.M{"group_id": groupId})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var results []*db_models.CliUserGroupMapping
	for cur.Next(ctx) {
		var m mUserGroupMapping
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMUserGroupMapping(&m))
	}
	return results, cur.Err()
}

func (c *Client) DeleteAllUserGroupMappingByUserId(userId int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colUserGroupMapping).DeleteMany(ctx, bson.M{"user_id": userId})
	return err
}

func (c *Client) DeleteAllUserGroupMappingByGroupId(groupId int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colUserGroupMapping).DeleteMany(ctx, bson.M{"group_id": groupId})
	return err
}

// ── cli_group_ne_mapping ─────────────────────────────────────────────────────

type mGroupNeMapping struct {
	GroupID int64 `bson:"group_id"`
	TblNeID int64 `bson:"tbl_ne_id"`
}

func fromMGroupNeMapping(m *mGroupNeMapping) *db_models.CliGroupNeMapping {
	return &db_models.CliGroupNeMapping{GroupID: m.GroupID, TblNeID: m.TblNeID}
}

func (c *Client) CreateGroupNeMapping(m *db_models.CliGroupNeMapping) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colGroupNeMapping).InsertOne(ctx, bson.M{"group_id": m.GroupID, "tbl_ne_id": m.TblNeID})
	return err
}

func (c *Client) DeleteGroupNeMapping(m *db_models.CliGroupNeMapping) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colGroupNeMapping).DeleteOne(ctx, bson.M{"group_id": m.GroupID, "tbl_ne_id": m.TblNeID})
	return err
}

func (c *Client) GetAllNesOfGroup(groupId int64) ([]*db_models.CliGroupNeMapping, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colGroupNeMapping).Find(ctx, bson.M{"group_id": groupId})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var results []*db_models.CliGroupNeMapping
	for cur.Next(ctx) {
		var m mGroupNeMapping
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMGroupNeMapping(&m))
	}
	return results, cur.Err()
}

func (c *Client) GetAllGroupsOfNe(neId int64) ([]*db_models.CliGroupNeMapping, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colGroupNeMapping).Find(ctx, bson.M{"tbl_ne_id": neId})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var results []*db_models.CliGroupNeMapping
	for cur.Next(ctx) {
		var m mGroupNeMapping
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		results = append(results, fromMGroupNeMapping(&m))
	}
	return results, cur.Err()
}

func (c *Client) DeleteAllGroupNeMappingByGroupId(groupId int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colGroupNeMapping).DeleteMany(ctx, bson.M{"group_id": groupId})
	return err
}

func (c *Client) DeleteAllGroupNeMappingByNeId(neId int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colGroupNeMapping).DeleteMany(ctx, bson.M{"tbl_ne_id": neId})
	return err
}
