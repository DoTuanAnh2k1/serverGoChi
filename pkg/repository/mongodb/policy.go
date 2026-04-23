package mongodb

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	colPasswordPolicy  = "cli_password_policy"
	colPasswordHistory = "cli_password_history"
	colMgtPermission   = "cli_group_mgt_permission"
)

// ───────── cli_password_policy ─────────

type mPasswordPolicy struct {
	ID               int64  `bson:"id"`
	Name             string `bson:"name"`
	MaxAgeDays       int32  `bson:"max_age_days"`
	MinLength        int32  `bson:"min_length"`
	RequireUppercase bool   `bson:"require_uppercase"`
	RequireLowercase bool   `bson:"require_lowercase"`
	RequireDigit     bool   `bson:"require_digit"`
	RequireSpecial   bool   `bson:"require_special"`
	HistoryCount     int32  `bson:"history_count"`
	MaxLoginFailure  int32  `bson:"max_login_failure"`
	LockoutMinutes   int32  `bson:"lockout_minutes"`
}

func toMPasswordPolicy(p *db_models.CliPasswordPolicy) *mPasswordPolicy {
	return &mPasswordPolicy{
		ID: p.ID, Name: p.Name,
		MaxAgeDays: p.MaxAgeDays, MinLength: p.MinLength,
		RequireUppercase: p.RequireUppercase, RequireLowercase: p.RequireLowercase,
		RequireDigit: p.RequireDigit, RequireSpecial: p.RequireSpecial,
		HistoryCount: p.HistoryCount, MaxLoginFailure: p.MaxLoginFailure,
		LockoutMinutes: p.LockoutMinutes,
	}
}

func fromMPasswordPolicy(m *mPasswordPolicy) *db_models.CliPasswordPolicy {
	return &db_models.CliPasswordPolicy{
		ID: m.ID, Name: m.Name,
		MaxAgeDays: m.MaxAgeDays, MinLength: m.MinLength,
		RequireUppercase: m.RequireUppercase, RequireLowercase: m.RequireLowercase,
		RequireDigit: m.RequireDigit, RequireSpecial: m.RequireSpecial,
		HistoryCount: m.HistoryCount, MaxLoginFailure: m.MaxLoginFailure,
		LockoutMinutes: m.LockoutMinutes,
	}
}

func (c *Client) CreatePasswordPolicy(p *db_models.CliPasswordPolicy) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if p.ID == 0 {
		id, err := c.nextID(ctx, colPasswordPolicy)
		if err != nil {
			return err
		}
		p.ID = id
	}
	_, err := c.col(colPasswordPolicy).InsertOne(ctx, toMPasswordPolicy(p))
	return err
}

func (c *Client) GetPasswordPolicyById(id int64) (*db_models.CliPasswordPolicy, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mPasswordPolicy
	err := c.col(colPasswordPolicy).FindOne(ctx, bson.M{"id": id}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMPasswordPolicy(&m), nil
}

func (c *Client) GetPasswordPolicyByName(name string) (*db_models.CliPasswordPolicy, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	var m mPasswordPolicy
	err := c.col(colPasswordPolicy).FindOne(ctx, bson.M{"name": name}).Decode(&m)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return fromMPasswordPolicy(&m), nil
}

func (c *Client) ListPasswordPolicies() ([]*db_models.CliPasswordPolicy, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colPasswordPolicy).Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*db_models.CliPasswordPolicy
	for cur.Next(ctx) {
		var m mPasswordPolicy
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, fromMPasswordPolicy(&m))
	}
	return out, cur.Err()
}

func (c *Client) UpdatePasswordPolicy(p *db_models.CliPasswordPolicy) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colPasswordPolicy).UpdateOne(ctx, bson.M{"id": p.ID}, bson.M{"$set": toMPasswordPolicy(p)})
	return err
}

func (c *Client) DeletePasswordPolicyById(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colPasswordPolicy).DeleteOne(ctx, bson.M{"id": id})
	return err
}

// ───────── cli_password_history ─────────

type mPasswordHistory struct {
	ID           int64     `bson:"id"`
	UserID       int64     `bson:"user_id"`
	PasswordHash string    `bson:"password_hash"`
	ChangedAt    time.Time `bson:"changed_at"`
}

func (c *Client) AppendPasswordHistory(h *db_models.CliPasswordHistory) error {
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
	_, err := c.col(colPasswordHistory).InsertOne(ctx, &mPasswordHistory{
		ID: h.ID, UserID: h.UserID, PasswordHash: h.PasswordHash, ChangedAt: h.ChangedAt,
	})
	return err
}

func (c *Client) GetRecentPasswordHistory(userID int64, limit int) ([]*db_models.CliPasswordHistory, error) {
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
	var out []*db_models.CliPasswordHistory
	for cur.Next(ctx) {
		var m mPasswordHistory
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, &db_models.CliPasswordHistory{
			ID: m.ID, UserID: m.UserID, PasswordHash: m.PasswordHash, ChangedAt: m.ChangedAt,
		})
	}
	return out, cur.Err()
}

// PrunePasswordHistory keeps only the most-recent `keep` rows per user.
// Two-phase: list the ids we want to keep, then delete the rest. The list
// is bounded (keep is small, ~12) so this is cheap.
func (c *Client) PrunePasswordHistory(userID int64, keep int) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if keep <= 0 {
		_, err := c.col(colPasswordHistory).DeleteMany(ctx, bson.M{"user_id": userID})
		return err
	}
	opts := options.Find().SetSort(bson.D{{Key: "changed_at", Value: -1}}).SetLimit(int64(keep)).SetProjection(bson.M{"id": 1})
	cur, err := c.col(colPasswordHistory).Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return err
	}
	var keepIDs []int64
	for cur.Next(ctx) {
		var row struct {
			ID int64 `bson:"id"`
		}
		if err := cur.Decode(&row); err != nil {
			cur.Close(ctx)
			return err
		}
		keepIDs = append(keepIDs, row.ID)
	}
	cur.Close(ctx)
	filter := bson.M{"user_id": userID}
	if len(keepIDs) > 0 {
		filter["id"] = bson.M{"$nin": keepIDs}
	}
	_, err = c.col(colPasswordHistory).DeleteMany(ctx, filter)
	return err
}

// ───────── cli_group_mgt_permission ─────────

type mMgtPermission struct {
	ID       int64  `bson:"id"`
	GroupID  int64  `bson:"group_id"`
	Resource string `bson:"resource"`
	Action   string `bson:"action"`
}

func (c *Client) CreateMgtPermission(p *db_models.CliGroupMgtPermission) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	if p.ID == 0 {
		id, err := c.nextID(ctx, colMgtPermission)
		if err != nil {
			return err
		}
		p.ID = id
	}
	_, err := c.col(colMgtPermission).InsertOne(ctx, &mMgtPermission{
		ID: p.ID, GroupID: p.GroupID, Resource: p.Resource, Action: p.Action,
	})
	return err
}

func (c *Client) ListMgtPermissions(groupID int64) ([]*db_models.CliGroupMgtPermission, error) {
	ctx, cancel := c.opCtx()
	defer cancel()
	cur, err := c.col(colMgtPermission).Find(ctx, bson.M{"group_id": groupID})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*db_models.CliGroupMgtPermission
	for cur.Next(ctx) {
		var m mMgtPermission
		if err := cur.Decode(&m); err != nil {
			return nil, err
		}
		out = append(out, &db_models.CliGroupMgtPermission{
			ID: m.ID, GroupID: m.GroupID, Resource: m.Resource, Action: m.Action,
		})
	}
	return out, cur.Err()
}

func (c *Client) DeleteMgtPermissionById(id int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colMgtPermission).DeleteOne(ctx, bson.M{"id": id})
	return err
}

func (c *Client) DeleteAllMgtPermissionByGroupId(groupID int64) error {
	ctx, cancel := c.opCtx()
	defer cancel()
	_, err := c.col(colMgtPermission).DeleteMany(ctx, bson.M{"group_id": groupID})
	return err
}
