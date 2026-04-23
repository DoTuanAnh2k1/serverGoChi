package mongodb

import (
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	colNeProfile            = "cli_ne_profile"
	colCommandDef           = "cli_command_def"
	colCommandGroup         = "cli_command_group"
	colCommandGroupMapping  = "cli_command_group_mapping"
	colGroupCmdPermission   = "cli_group_cmd_permission"
)

// ───────── cli_ne_profile ─────────

type mNeProfile struct {
	ID          int64  `bson:"id"`
	Name        string `bson:"name"`
	Description string `bson:"description"`
}

func toMNeProfile(p *db_models.CliNeProfile) *mNeProfile {
	return &mNeProfile{ID: p.ID, Name: p.Name, Description: p.Description}
}

func fromMNeProfile(m *mNeProfile) *db_models.CliNeProfile {
	return &db_models.CliNeProfile{ID: m.ID, Name: m.Name, Description: m.Description}
}

func (c *Client) CreateNeProfile(p *db_models.CliNeProfile) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if p.ID == 0 {
		id, err := c.nextID(ctx, colNeProfile)
		if err != nil {
			return err
		}
		p.ID = id
	}
	_, err := c.col(colNeProfile).InsertOne(ctx, toMNeProfile(p))
	return err
}

func (c *Client) GetNeProfileById(id int64) (*db_models.CliNeProfile, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mNeProfile
	err := c.col(colNeProfile).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMNeProfile(&m), nil
}

func (c *Client) GetNeProfileByName(name string) (*db_models.CliNeProfile, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mNeProfile
	err := c.col(colNeProfile).FindOne(ctx, bson.M{"name": name}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMNeProfile(&m), nil
}

func (c *Client) ListNeProfiles() ([]*db_models.CliNeProfile, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colNeProfile).Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*db_models.CliNeProfile
	for cur.Next(ctx) {
		var m mNeProfile
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, fromMNeProfile(&m))
	}
	return out, cur.Err()
}

func (c *Client) UpdateNeProfile(p *db_models.CliNeProfile) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colNeProfile).UpdateOne(ctx, bson.M{"id": p.ID}, bson.M{"$set": toMNeProfile(p)})
	return err
}

func (c *Client) DeleteNeProfileById(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colNeProfile).DeleteOne(ctx, bson.M{"id": id})
	return err
}

// ───────── cli_command_def ─────────

type mCommandDef struct {
	ID          int64  `bson:"id"`
	Service     string `bson:"service"`
	NeProfile   string `bson:"ne_profile"`
	Pattern     string `bson:"pattern"`
	Category    string `bson:"category"`
	RiskLevel   int32  `bson:"risk_level"`
	Description string `bson:"description"`
	CreatedBy   string `bson:"created_by"`
}

func toMCommandDef(d *db_models.CliCommandDef) *mCommandDef {
	return &mCommandDef{
		ID:          d.ID,
		Service:     d.Service,
		NeProfile:   d.NeProfile,
		Pattern:     d.Pattern,
		Category:    d.Category,
		RiskLevel:   d.RiskLevel,
		Description: d.Description,
		CreatedBy:   d.CreatedBy,
	}
}

func fromMCommandDef(m *mCommandDef) *db_models.CliCommandDef {
	return &db_models.CliCommandDef{
		ID:          m.ID,
		Service:     m.Service,
		NeProfile:   m.NeProfile,
		Pattern:     m.Pattern,
		Category:    m.Category,
		RiskLevel:   m.RiskLevel,
		Description: m.Description,
		CreatedBy:   m.CreatedBy,
	}
}

func (c *Client) CreateCommandDef(d *db_models.CliCommandDef) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if d.ID == 0 {
		id, err := c.nextID(ctx, colCommandDef)
		if err != nil {
			return err
		}
		d.ID = id
	}
	_, err := c.col(colCommandDef).InsertOne(ctx, toMCommandDef(d))
	return err
}

func (c *Client) GetCommandDefById(id int64) (*db_models.CliCommandDef, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mCommandDef
	err := c.col(colCommandDef).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMCommandDef(&m), nil
}

func (c *Client) ListCommandDefs(service, neProfile, category string) ([]*db_models.CliCommandDef, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	filter := bson.M{}
	if service != "" {
		filter["service"] = service
	}
	if neProfile != "" {
		filter["ne_profile"] = neProfile
	}
	if category != "" {
		filter["category"] = category
	}
	cur, err := c.col(colCommandDef).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*db_models.CliCommandDef
	for cur.Next(ctx) {
		var m mCommandDef
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, fromMCommandDef(&m))
	}
	return out, cur.Err()
}

func (c *Client) UpdateCommandDef(d *db_models.CliCommandDef) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCommandDef).UpdateOne(ctx, bson.M{"id": d.ID}, bson.M{"$set": toMCommandDef(d)})
	return err
}

func (c *Client) DeleteCommandDefById(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCommandDef).DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	// Cascade: remove from any command groups.
	_, _ = c.col(colCommandGroupMapping).DeleteMany(ctx, bson.M{"command_def_id": id})
	return nil
}

// ───────── cli_command_group ─────────

type mCommandGroup struct {
	ID          int64  `bson:"id"`
	Name        string `bson:"name"`
	NeProfile   string `bson:"ne_profile"`
	Service     string `bson:"service"`
	Description string `bson:"description"`
	CreatedBy   string `bson:"created_by"`
}

func toMCommandGroup(g *db_models.CliCommandGroup) *mCommandGroup {
	return &mCommandGroup{
		ID:          g.ID,
		Name:        g.Name,
		NeProfile:   g.NeProfile,
		Service:     g.Service,
		Description: g.Description,
		CreatedBy:   g.CreatedBy,
	}
}

func fromMCommandGroup(m *mCommandGroup) *db_models.CliCommandGroup {
	return &db_models.CliCommandGroup{
		ID:          m.ID,
		Name:        m.Name,
		NeProfile:   m.NeProfile,
		Service:     m.Service,
		Description: m.Description,
		CreatedBy:   m.CreatedBy,
	}
}

func (c *Client) CreateCommandGroup(g *db_models.CliCommandGroup) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if g.ID == 0 {
		id, err := c.nextID(ctx, colCommandGroup)
		if err != nil {
			return err
		}
		g.ID = id
	}
	_, err := c.col(colCommandGroup).InsertOne(ctx, toMCommandGroup(g))
	return err
}

func (c *Client) GetCommandGroupById(id int64) (*db_models.CliCommandGroup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mCommandGroup
	err := c.col(colCommandGroup).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMCommandGroup(&m), nil
}

func (c *Client) GetCommandGroupByName(name string) (*db_models.CliCommandGroup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mCommandGroup
	err := c.col(colCommandGroup).FindOne(ctx, bson.M{"name": name}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMCommandGroup(&m), nil
}

func (c *Client) ListCommandGroups(service, neProfile string) ([]*db_models.CliCommandGroup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	filter := bson.M{}
	if service != "" {
		filter["service"] = service
	}
	if neProfile != "" {
		filter["ne_profile"] = neProfile
	}
	cur, err := c.col(colCommandGroup).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*db_models.CliCommandGroup
	for cur.Next(ctx) {
		var m mCommandGroup
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, fromMCommandGroup(&m))
	}
	return out, cur.Err()
}

func (c *Client) UpdateCommandGroup(g *db_models.CliCommandGroup) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCommandGroup).UpdateOne(ctx, bson.M{"id": g.ID}, bson.M{"$set": toMCommandGroup(g)})
	return err
}

func (c *Client) DeleteCommandGroupById(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCommandGroup).DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}
	_, _ = c.col(colCommandGroupMapping).DeleteMany(ctx, bson.M{"command_group_id": id})
	return nil
}

// ───────── cli_command_group_mapping ─────────

func (c *Client) AddCommandToGroup(m *db_models.CliCommandGroupMapping) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCommandGroupMapping).InsertOne(ctx, bson.M{
		"command_group_id": m.CommandGroupID,
		"command_def_id":   m.CommandDefID,
	})
	return err
}

func (c *Client) RemoveCommandFromGroup(m *db_models.CliCommandGroupMapping) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCommandGroupMapping).DeleteOne(ctx, bson.M{
		"command_group_id": m.CommandGroupID,
		"command_def_id":   m.CommandDefID,
	})
	return err
}

func (c *Client) ListCommandsOfGroup(groupId int64) ([]*db_models.CliCommandDef, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colCommandGroupMapping).Find(ctx, bson.M{"command_group_id": groupId})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var defIds []int64
	for cur.Next(ctx) {
		var row struct {
			CommandDefID int64 `bson:"command_def_id"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		defIds = append(defIds, row.CommandDefID)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	if len(defIds) == 0 {
		return []*db_models.CliCommandDef{}, nil
	}
	defCur, err := c.col(colCommandDef).Find(ctx, bson.M{"id": bson.M{"$in": defIds}})
	if err != nil {
		return nil, err
	}
	defer defCur.Close(ctx)
	var out []*db_models.CliCommandDef
	for defCur.Next(ctx) {
		var m mCommandDef
		if err := defCur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, fromMCommandDef(&m))
	}
	return out, defCur.Err()
}

func (c *Client) ListGroupsOfCommand(commandId int64) ([]*db_models.CliCommandGroup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colCommandGroupMapping).Find(ctx, bson.M{"command_def_id": commandId})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var groupIds []int64
	for cur.Next(ctx) {
		var row struct {
			CommandGroupID int64 `bson:"command_group_id"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		groupIds = append(groupIds, row.CommandGroupID)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	if len(groupIds) == 0 {
		return []*db_models.CliCommandGroup{}, nil
	}
	grpCur, err := c.col(colCommandGroup).Find(ctx, bson.M{"id": bson.M{"$in": groupIds}})
	if err != nil {
		return nil, err
	}
	defer grpCur.Close(ctx)
	var out []*db_models.CliCommandGroup
	for grpCur.Next(ctx) {
		var m mCommandGroup
		if err := grpCur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, fromMCommandGroup(&m))
	}
	return out, grpCur.Err()
}

func (c *Client) DeleteAllCommandGroupMappingByGroupId(groupId int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCommandGroupMapping).DeleteMany(ctx, bson.M{"command_group_id": groupId})
	return err
}

func (c *Client) DeleteAllCommandGroupMappingByCommandId(commandId int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCommandGroupMapping).DeleteMany(ctx, bson.M{"command_def_id": commandId})
	return err
}

// ───────── cli_group_cmd_permission ─────────

type mGroupCmdPermission struct {
	ID         int64  `bson:"id"`
	GroupID    int64  `bson:"group_id"`
	Service    string `bson:"service"`
	NeScope    string `bson:"ne_scope"`
	GrantType  string `bson:"grant_type"`
	GrantValue string `bson:"grant_value"`
	Effect     string `bson:"effect"`
}

func toMGroupCmdPermission(p *db_models.CliGroupCmdPermission) *mGroupCmdPermission {
	return &mGroupCmdPermission{
		ID:         p.ID,
		GroupID:    p.GroupID,
		Service:    p.Service,
		NeScope:    p.NeScope,
		GrantType:  p.GrantType,
		GrantValue: p.GrantValue,
		Effect:     p.Effect,
	}
}

func fromMGroupCmdPermission(m *mGroupCmdPermission) *db_models.CliGroupCmdPermission {
	return &db_models.CliGroupCmdPermission{
		ID:         m.ID,
		GroupID:    m.GroupID,
		Service:    m.Service,
		NeScope:    m.NeScope,
		GrantType:  m.GrantType,
		GrantValue: m.GrantValue,
		Effect:     m.Effect,
	}
}

func (c *Client) CreateGroupCmdPermission(p *db_models.CliGroupCmdPermission) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if p.ID == 0 {
		id, err := c.nextID(ctx, colGroupCmdPermission)
		if err != nil {
			return err
		}
		p.ID = id
	}
	_, err := c.col(colGroupCmdPermission).InsertOne(ctx, toMGroupCmdPermission(p))
	return err
}

func (c *Client) GetGroupCmdPermissionById(id int64) (*db_models.CliGroupCmdPermission, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mGroupCmdPermission
	err := c.col(colGroupCmdPermission).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMGroupCmdPermission(&m), nil
}

func (c *Client) ListGroupCmdPermissions(groupId int64) ([]*db_models.CliGroupCmdPermission, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colGroupCmdPermission).Find(ctx, bson.M{"group_id": groupId})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*db_models.CliGroupCmdPermission
	for cur.Next(ctx) {
		var m mGroupCmdPermission
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, fromMGroupCmdPermission(&m))
	}
	return out, cur.Err()
}

func (c *Client) DeleteGroupCmdPermissionById(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colGroupCmdPermission).DeleteOne(ctx, bson.M{"id": id})
	return err
}

func (c *Client) DeleteAllGroupCmdPermissionByGroupId(groupId int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colGroupCmdPermission).DeleteMany(ctx, bson.M{"group_id": groupId})
	return err
}
