package mongodb

import (
	"errors"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func isNoDocs(err error) bool { return errors.Is(err, mongo.ErrNoDocuments) }

// ── User ────────────────────────────────────────────────────────────────

func (c *Client) CreateUser(u *db_models.User) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if u.ID == 0 {
		id, err := c.nextID(ctx, colUser)
		if err != nil {
			return err
		}
		u.ID = id
	}
	now := time.Now().UTC()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = now
	}
	_, err := c.col(colUser).InsertOne(ctx, toMUser(u))
	return err
}

func (c *Client) GetUserByID(id int64) (*db_models.User, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mUser
	err := c.col(colUser).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if isNoDocs(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMUser(&m), nil
}

func (c *Client) GetUserByUsername(username string) (*db_models.User, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mUser
	err := c.col(colUser).FindOne(ctx, bson.M{"username": username}).Decode(&m)
	if isNoDocs(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMUser(&m), nil
}

func (c *Client) ListUsers() ([]*db_models.User, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colUser).Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*db_models.User
	for cur.Next(ctx) {
		var m mUser
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, fromMUser(&m))
	}
	return out, cur.Err()
}

func (c *Client) UpdateUser(u *db_models.User) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	u.UpdatedAt = time.Now().UTC()
	_, err := c.col(colUser).ReplaceOne(ctx, bson.M{"id": u.ID}, toMUser(u))
	return err
}

func (c *Client) DeleteUserByID(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colUser).DeleteOne(ctx, bson.M{"id": id})
	return err
}

// ── NE ──────────────────────────────────────────────────────────────────

func (c *Client) CreateNE(n *db_models.NE) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if n.ID == 0 {
		id, err := c.nextID(ctx, colNe)
		if err != nil {
			return err
		}
		n.ID = id
	}
	_, err := c.col(colNe).InsertOne(ctx, toMNE(n))
	return err
}

func (c *Client) GetNEByID(id int64) (*db_models.NE, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mNE
	err := c.col(colNe).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if isNoDocs(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMNE(&m), nil
}

func (c *Client) GetNEByNamespace(ns string) (*db_models.NE, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mNE
	err := c.col(colNe).FindOne(ctx, bson.M{"namespace": ns}).Decode(&m)
	if isNoDocs(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMNE(&m), nil
}

func (c *Client) ListNEs() ([]*db_models.NE, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colNe).Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*db_models.NE
	for cur.Next(ctx) {
		var m mNE
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, fromMNE(&m))
	}
	return out, cur.Err()
}

func (c *Client) UpdateNE(n *db_models.NE) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colNe).ReplaceOne(ctx, bson.M{"id": n.ID}, toMNE(n))
	return err
}

func (c *Client) DeleteNEByID(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colNe).DeleteOne(ctx, bson.M{"id": id})
	return err
}

// ── Command ─────────────────────────────────────────────────────────────

func (c *Client) CreateCommand(cmd *db_models.Command) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if cmd.ID == 0 {
		id, err := c.nextID(ctx, colCommand)
		if err != nil {
			return err
		}
		cmd.ID = id
	}
	_, err := c.col(colCommand).InsertOne(ctx, toMCommand(cmd))
	return err
}

func (c *Client) GetCommandByID(id int64) (*db_models.Command, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mCommand
	err := c.col(colCommand).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if isNoDocs(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMCommand(&m), nil
}

func (c *Client) GetCommandByTriple(neID int64, service, cmdText string) (*db_models.Command, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mCommand
	err := c.col(colCommand).FindOne(ctx, bson.M{
		"ne_id": neID, "service": service, "cmd_text": cmdText,
	}).Decode(&m)
	if isNoDocs(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMCommand(&m), nil
}

func (c *Client) ListCommands(neID int64, service string) ([]*db_models.Command, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	filter := bson.M{}
	if neID > 0 {
		filter["ne_id"] = neID
	}
	if service != "" {
		filter["service"] = service
	}
	opts := options.Find().SetSort(bson.D{{Key: "ne_id", Value: 1}, {Key: "service", Value: 1}, {Key: "cmd_text", Value: 1}})
	cur, err := c.col(colCommand).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*db_models.Command
	for cur.Next(ctx) {
		var m mCommand
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, fromMCommand(&m))
	}
	return out, cur.Err()
}

func (c *Client) UpdateCommand(cmd *db_models.Command) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCommand).ReplaceOne(ctx, bson.M{"id": cmd.ID}, toMCommand(cmd))
	return err
}

func (c *Client) DeleteCommandByID(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCommand).DeleteOne(ctx, bson.M{"id": id})
	return err
}

// ── NE Access Group ─────────────────────────────────────────────────────

func (c *Client) CreateNeAccessGroup(g *db_models.NeAccessGroup) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if g.ID == 0 {
		id, err := c.nextID(ctx, colNeAccessGroup)
		if err != nil {
			return err
		}
		g.ID = id
	}
	_, err := c.col(colNeAccessGroup).InsertOne(ctx, mGroup{ID: g.ID, Name: g.Name, Description: g.Description})
	return err
}

func (c *Client) GetNeAccessGroupByID(id int64) (*db_models.NeAccessGroup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mGroup
	err := c.col(colNeAccessGroup).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if isNoDocs(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &db_models.NeAccessGroup{ID: m.ID, Name: m.Name, Description: m.Description}, nil
}

func (c *Client) GetNeAccessGroupByName(name string) (*db_models.NeAccessGroup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mGroup
	err := c.col(colNeAccessGroup).FindOne(ctx, bson.M{"name": name}).Decode(&m)
	if isNoDocs(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &db_models.NeAccessGroup{ID: m.ID, Name: m.Name, Description: m.Description}, nil
}

func (c *Client) ListNeAccessGroups() ([]*db_models.NeAccessGroup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colNeAccessGroup).Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*db_models.NeAccessGroup
	for cur.Next(ctx) {
		var m mGroup
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, &db_models.NeAccessGroup{ID: m.ID, Name: m.Name, Description: m.Description})
	}
	return out, cur.Err()
}

func (c *Client) UpdateNeAccessGroup(g *db_models.NeAccessGroup) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colNeAccessGroup).ReplaceOne(ctx, bson.M{"id": g.ID},
		mGroup{ID: g.ID, Name: g.Name, Description: g.Description})
	return err
}

func (c *Client) DeleteNeAccessGroupByID(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if _, err := c.col(colNeAccessGroup).DeleteOne(ctx, bson.M{"id": id}); err != nil {
		return err
	}
	if _, err := c.col(colNeAccessGroupUser).DeleteMany(ctx, bson.M{"group_id": id}); err != nil {
		return err
	}
	_, err := c.col(colNeAccessGroupNe).DeleteMany(ctx, bson.M{"group_id": id})
	return err
}

func (c *Client) AddUserToNeAccessGroup(groupID, userID int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colNeAccessGroupUser).InsertOne(ctx, bson.M{"group_id": groupID, "user_id": userID})
	return err
}

func (c *Client) RemoveUserFromNeAccessGroup(groupID, userID int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colNeAccessGroupUser).DeleteOne(ctx, bson.M{"group_id": groupID, "user_id": userID})
	return err
}

func (c *Client) ListUsersInNeAccessGroup(groupID int64) ([]int64, error) {
	return c.distinctInt64(colNeAccessGroupUser, "user_id", bson.M{"group_id": groupID})
}

func (c *Client) ListNeAccessGroupsOfUser(userID int64) ([]int64, error) {
	return c.distinctInt64(colNeAccessGroupUser, "group_id", bson.M{"user_id": userID})
}

func (c *Client) AddNeToNeAccessGroup(groupID, neID int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colNeAccessGroupNe).InsertOne(ctx, bson.M{"group_id": groupID, "ne_id": neID})
	return err
}

func (c *Client) RemoveNeFromNeAccessGroup(groupID, neID int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colNeAccessGroupNe).DeleteOne(ctx, bson.M{"group_id": groupID, "ne_id": neID})
	return err
}

func (c *Client) ListNEsInNeAccessGroup(groupID int64) ([]int64, error) {
	return c.distinctInt64(colNeAccessGroupNe, "ne_id", bson.M{"group_id": groupID})
}

func (c *Client) ListNeAccessGroupsOfNE(neID int64) ([]int64, error) {
	return c.distinctInt64(colNeAccessGroupNe, "group_id", bson.M{"ne_id": neID})
}

// ── Cmd Exec Group ──────────────────────────────────────────────────────

func (c *Client) CreateCmdExecGroup(g *db_models.CmdExecGroup) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if g.ID == 0 {
		id, err := c.nextID(ctx, colCmdExecGroup)
		if err != nil {
			return err
		}
		g.ID = id
	}
	_, err := c.col(colCmdExecGroup).InsertOne(ctx, mGroup{ID: g.ID, Name: g.Name, Description: g.Description})
	return err
}

func (c *Client) GetCmdExecGroupByID(id int64) (*db_models.CmdExecGroup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mGroup
	err := c.col(colCmdExecGroup).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if isNoDocs(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &db_models.CmdExecGroup{ID: m.ID, Name: m.Name, Description: m.Description}, nil
}

func (c *Client) GetCmdExecGroupByName(name string) (*db_models.CmdExecGroup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mGroup
	err := c.col(colCmdExecGroup).FindOne(ctx, bson.M{"name": name}).Decode(&m)
	if isNoDocs(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &db_models.CmdExecGroup{ID: m.ID, Name: m.Name, Description: m.Description}, nil
}

func (c *Client) ListCmdExecGroups() ([]*db_models.CmdExecGroup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colCmdExecGroup).Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*db_models.CmdExecGroup
	for cur.Next(ctx) {
		var m mGroup
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, &db_models.CmdExecGroup{ID: m.ID, Name: m.Name, Description: m.Description})
	}
	return out, cur.Err()
}

func (c *Client) UpdateCmdExecGroup(g *db_models.CmdExecGroup) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCmdExecGroup).ReplaceOne(ctx, bson.M{"id": g.ID},
		mGroup{ID: g.ID, Name: g.Name, Description: g.Description})
	return err
}

func (c *Client) DeleteCmdExecGroupByID(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if _, err := c.col(colCmdExecGroup).DeleteOne(ctx, bson.M{"id": id}); err != nil {
		return err
	}
	if _, err := c.col(colCmdExecGroupUser).DeleteMany(ctx, bson.M{"group_id": id}); err != nil {
		return err
	}
	_, err := c.col(colCmdExecGroupCommand).DeleteMany(ctx, bson.M{"group_id": id})
	return err
}

func (c *Client) AddUserToCmdExecGroup(groupID, userID int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCmdExecGroupUser).InsertOne(ctx, bson.M{"group_id": groupID, "user_id": userID})
	return err
}

func (c *Client) RemoveUserFromCmdExecGroup(groupID, userID int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCmdExecGroupUser).DeleteOne(ctx, bson.M{"group_id": groupID, "user_id": userID})
	return err
}

func (c *Client) ListUsersInCmdExecGroup(groupID int64) ([]int64, error) {
	return c.distinctInt64(colCmdExecGroupUser, "user_id", bson.M{"group_id": groupID})
}

func (c *Client) ListCmdExecGroupsOfUser(userID int64) ([]int64, error) {
	return c.distinctInt64(colCmdExecGroupUser, "group_id", bson.M{"user_id": userID})
}

func (c *Client) AddCommandToCmdExecGroup(groupID, commandID int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCmdExecGroupCommand).InsertOne(ctx, bson.M{"group_id": groupID, "command_id": commandID})
	return err
}

func (c *Client) RemoveCommandFromCmdExecGroup(groupID, commandID int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colCmdExecGroupCommand).DeleteOne(ctx, bson.M{"group_id": groupID, "command_id": commandID})
	return err
}

func (c *Client) ListCommandsInCmdExecGroup(groupID int64) ([]int64, error) {
	return c.distinctInt64(colCmdExecGroupCommand, "command_id", bson.M{"group_id": groupID})
}

func (c *Client) ListCmdExecGroupsOfCommand(commandID int64) ([]int64, error) {
	return c.distinctInt64(colCmdExecGroupCommand, "group_id", bson.M{"command_id": commandID})
}

// ── Password Policy / History ───────────────────────────────────────────

func (c *Client) GetPasswordPolicy() (*db_models.PasswordPolicy, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mPasswordPolicy
	err := c.col(colPasswordPolicy).FindOne(ctx, bson.M{"id": int64(1)}).Decode(&m)
	if isNoDocs(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMPasswordPolicy(&m), nil
}

func (c *Client) UpsertPasswordPolicy(p *db_models.PasswordPolicy) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	p.ID = 1
	_, err := c.col(colPasswordPolicy).ReplaceOne(ctx,
		bson.M{"id": int64(1)}, toMPasswordPolicy(p),
		options.Replace().SetUpsert(true))
	return err
}

func (c *Client) AppendPasswordHistory(h *db_models.PasswordHistory) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if h.ID == 0 {
		id, err := c.nextID(ctx, colPasswordHistory)
		if err != nil {
			return err
		}
		h.ID = id
	}
	if h.ChangedAt.IsZero() {
		h.ChangedAt = time.Now().UTC()
	}
	_, err := c.col(colPasswordHistory).InsertOne(ctx, mPasswordHistory{
		ID: h.ID, UserID: h.UserID, PasswordHash: h.PasswordHash, ChangedAt: h.ChangedAt,
	})
	return err
}

func (c *Client) GetRecentPasswordHistory(userID int64, limit int) ([]*db_models.PasswordHistory, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	opts := options.Find().SetSort(bson.D{{Key: "changed_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cur, err := c.col(colPasswordHistory).Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*db_models.PasswordHistory
	for cur.Next(ctx) {
		var m mPasswordHistory
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, &db_models.PasswordHistory{
			ID: m.ID, UserID: m.UserID, PasswordHash: m.PasswordHash, ChangedAt: m.ChangedAt,
		})
	}
	return out, cur.Err()
}

func (c *Client) PrunePasswordHistory(userID int64, keep int) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if keep <= 0 {
		_, err := c.col(colPasswordHistory).DeleteMany(ctx, bson.M{"user_id": userID})
		return err
	}
	opts := options.Find().
		SetSort(bson.D{{Key: "changed_at", Value: -1}}).
		SetLimit(int64(keep)).
		SetProjection(bson.M{"id": 1})
	cur, err := c.col(colPasswordHistory).Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return err
	}
	defer cur.Close(ctx)
	var keepIDs []int64
	for cur.Next(ctx) {
		var doc struct {
			ID int64 `bson:"id"`
		}
		if err := cur.Decode(&doc); err != nil {
			return err
		}
		keepIDs = append(keepIDs, doc.ID)
	}
	if err := cur.Err(); err != nil {
		return err
	}
	filter := bson.M{"user_id": userID}
	if len(keepIDs) > 0 {
		filter["id"] = bson.M{"$nin": keepIDs}
	}
	_, err = c.col(colPasswordHistory).DeleteMany(ctx, filter)
	return err
}

// ── User Access List ────────────────────────────────────────────────────

func (c *Client) CreateAccessListEntry(e *db_models.UserAccessList) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if e.ID == 0 {
		id, err := c.nextID(ctx, colUserAccessList)
		if err != nil {
			return err
		}
		e.ID = id
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	_, err := c.col(colUserAccessList).InsertOne(ctx, mUserAccessList{
		ID: e.ID, ListType: e.ListType, MatchType: e.MatchType,
		Pattern: e.Pattern, Reason: e.Reason, CreatedAt: e.CreatedAt,
	})
	return err
}

func (c *Client) ListAccessListEntries(listType string) ([]*db_models.UserAccessList, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	filter := bson.M{}
	if listType != "" {
		filter["list_type"] = listType
	}
	cur, err := c.col(colUserAccessList).Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*db_models.UserAccessList
	for cur.Next(ctx) {
		var m mUserAccessList
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, &db_models.UserAccessList{
			ID: m.ID, ListType: m.ListType, MatchType: m.MatchType,
			Pattern: m.Pattern, Reason: m.Reason, CreatedAt: m.CreatedAt,
		})
	}
	return out, cur.Err()
}

func (c *Client) DeleteAccessListEntryByID(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colUserAccessList).DeleteOne(ctx, bson.M{"id": id})
	return err
}

// ── History ─────────────────────────────────────────────────────────────

func (c *Client) SaveOperationHistory(h db_models.OperationHistory) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if h.ID == 0 {
		id, err := c.nextID(ctx, colOperationHistory)
		if err != nil {
			return err
		}
		h.ID = int32(id)
	}
	if h.CreatedDate.IsZero() {
		h.CreatedDate = time.Now().UTC()
	}
	_, err := c.col(colOperationHistory).InsertOne(ctx, toMOperationHistory(h))
	return err
}

func (c *Client) GetRecentHistory(limit int) ([]db_models.OperationHistory, error) {
	return c.GetRecentHistoryFiltered(limit, "", "", "")
}

func (c *Client) GetRecentHistoryFiltered(limit int, scope, neNamespace, account string) ([]db_models.OperationHistory, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	filter := bson.M{}
	if scope != "" {
		filter["scope"] = scope
	}
	if neNamespace != "" {
		filter["ne_namespace"] = neNamespace
	}
	if account != "" {
		filter["account"] = account
	}
	opts := options.Find().SetSort(bson.D{{Key: "created_date", Value: -1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cur, err := c.col(colOperationHistory).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []db_models.OperationHistory
	for cur.Next(ctx) {
		var m mOperationHistory
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, fromMOperationHistory(&m))
	}
	return out, cur.Err()
}

func (c *Client) GetDailyOperationHistory(date time.Time) ([]db_models.OperationHistory, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)
	filter := bson.M{"created_date": bson.M{"$gte": start, "$lt": end}}
	opts := options.Find().SetSort(bson.D{{Key: "ne_namespace", Value: 1}, {Key: "created_date", Value: 1}})
	cur, err := c.col(colOperationHistory).Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []db_models.OperationHistory
	for cur.Next(ctx) {
		var m mOperationHistory
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, fromMOperationHistory(&m))
	}
	return out, cur.Err()
}

func (c *Client) DeleteHistoryBefore(cutoff time.Time) (int64, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	res, err := c.col(colOperationHistory).DeleteMany(ctx, bson.M{"created_date": bson.M{"$lt": cutoff}})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

func (c *Client) UpdateLoginHistory(username, ip string, t time.Time) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	id, err := c.nextID(ctx, colLoginHistory)
	if err != nil {
		return err
	}
	_, err = c.col(colLoginHistory).InsertOne(ctx, mLoginHistory{
		ID: int32(id), Username: username, IPAddress: ip, TimeLogin: t,
	})
	return err
}

// ── Config Backup ───────────────────────────────────────────────────────

func (c *Client) SaveConfigBackup(b *db_models.ConfigBackup) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if b.ID == 0 {
		id, err := c.nextID(ctx, colConfigBackup)
		if err != nil {
			return err
		}
		b.ID = id
	}
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now().UTC()
	}
	_, err := c.col(colConfigBackup).InsertOne(ctx, toMConfigBackup(b))
	return err
}

func (c *Client) ListConfigBackups(neName string) ([]*db_models.ConfigBackup, error) {
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
	var out []*db_models.ConfigBackup
	for cur.Next(ctx) {
		var m mConfigBackup
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, fromMConfigBackup(&m))
	}
	return out, cur.Err()
}

func (c *Client) GetConfigBackupByID(id int64) (*db_models.ConfigBackup, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mConfigBackup
	err := c.col(colConfigBackup).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if isNoDocs(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMConfigBackup(&m), nil
}

// distinctInt64 is a small helper for pivot-table lookups like "give me all
// user_ids in this group". Using Distinct keeps us server-side and avoids
// streaming every pivot row just to pluck one field.
func (c *Client) distinctInt64(coll, field string, filter bson.M) ([]int64, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	raw, err := c.col(coll).Distinct(ctx, field, filter)
	if err != nil {
		return nil, err
	}
	out := make([]int64, 0, len(raw))
	for _, v := range raw {
		switch n := v.(type) {
		case int64:
			out = append(out, n)
		case int32:
			out = append(out, int64(n))
		case int:
			out = append(out, int64(n))
		}
	}
	return out, nil
}
