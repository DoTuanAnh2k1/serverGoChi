package mongodb

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// Mongo documents — one struct per collection, bson-tagged in snake_case so
// the shape matches db.sql column names. The native driver has no ORM, so
// converters below handle the db_models ↔ document mapping explicitly.

type mUser struct {
	ID                int64      `bson:"id"`
	Username          string     `bson:"username"`
	PasswordHash      string     `bson:"password_hash"`
	Email             string     `bson:"email"`
	FullName          string     `bson:"full_name"`
	Phone             string     `bson:"phone"`
	IsEnabled         bool       `bson:"is_enabled"`
	PasswordExpiresAt *time.Time `bson:"password_expires_at,omitempty"`
	LoginFailureCount int32      `bson:"login_failure_count"`
	LockedAt          *time.Time `bson:"locked_at,omitempty"`
	LastLoginAt       *time.Time `bson:"last_login_at,omitempty"`
	CreatedAt         time.Time  `bson:"created_at"`
	UpdatedAt         time.Time  `bson:"updated_at"`
}

func toMUser(u *db_models.User) mUser {
	return mUser{
		ID: u.ID, Username: u.Username, PasswordHash: u.PasswordHash,
		Email: u.Email, FullName: u.FullName, Phone: u.Phone,
		IsEnabled: u.IsEnabled, PasswordExpiresAt: u.PasswordExpiresAt,
		LoginFailureCount: u.LoginFailureCount, LockedAt: u.LockedAt,
		LastLoginAt: u.LastLoginAt, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
	}
}

func fromMUser(m *mUser) *db_models.User {
	return &db_models.User{
		ID: m.ID, Username: m.Username, PasswordHash: m.PasswordHash,
		Email: m.Email, FullName: m.FullName, Phone: m.Phone,
		IsEnabled: m.IsEnabled, PasswordExpiresAt: m.PasswordExpiresAt,
		LoginFailureCount: m.LoginFailureCount, LockedAt: m.LockedAt,
		LastLoginAt: m.LastLoginAt, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

type mNE struct {
	ID          int64  `bson:"id"`
	Namespace   string `bson:"namespace"`
	NeType      string `bson:"ne_type"`
	SiteName    string `bson:"site_name"`
	Description string `bson:"description"`
	MasterIP    string `bson:"master_ip"`
	MasterPort  int32  `bson:"master_port"`
	SSHUsername string `bson:"ssh_username"`
	SSHPassword string `bson:"ssh_password"`
	CommandURL  string `bson:"command_url"`
	ConfMode    string `bson:"conf_mode"`
}

func toMNE(n *db_models.NE) mNE {
	return mNE{
		ID: n.ID, Namespace: n.Namespace, NeType: n.NeType,
		SiteName: n.SiteName, Description: n.Description,
		MasterIP: n.MasterIP, MasterPort: n.MasterPort,
		SSHUsername: n.SSHUsername, SSHPassword: n.SSHPassword,
		CommandURL: n.CommandURL, ConfMode: n.ConfMode,
	}
}

func fromMNE(m *mNE) *db_models.NE {
	return &db_models.NE{
		ID: m.ID, Namespace: m.Namespace, NeType: m.NeType,
		SiteName: m.SiteName, Description: m.Description,
		MasterIP: m.MasterIP, MasterPort: m.MasterPort,
		SSHUsername: m.SSHUsername, SSHPassword: m.SSHPassword,
		CommandURL: m.CommandURL, ConfMode: m.ConfMode,
	}
}

type mCommand struct {
	ID          int64  `bson:"id"`
	NeID        int64  `bson:"ne_id"`
	Service     string `bson:"service"`
	CmdText     string `bson:"cmd_text"`
	Description string `bson:"description"`
}

func toMCommand(c *db_models.Command) mCommand {
	return mCommand{ID: c.ID, NeID: c.NeID, Service: c.Service, CmdText: c.CmdText, Description: c.Description}
}

func fromMCommand(m *mCommand) *db_models.Command {
	return &db_models.Command{ID: m.ID, NeID: m.NeID, Service: m.Service, CmdText: m.CmdText, Description: m.Description}
}

type mGroup struct {
	ID          int64  `bson:"id"`
	Name        string `bson:"name"`
	Description string `bson:"description"`
}

type mPasswordPolicy struct {
	ID                int64 `bson:"id"`
	MinLength         int32 `bson:"min_length"`
	MaxAgeDays        int32 `bson:"max_age_days"`
	RequireUppercase  bool  `bson:"require_uppercase"`
	RequireLowercase  bool  `bson:"require_lowercase"`
	RequireDigit      bool  `bson:"require_digit"`
	RequireSpecial    bool  `bson:"require_special"`
	HistoryCount      int32 `bson:"history_count"`
	MaxLoginFailure   int32 `bson:"max_login_failure"`
	LockoutMinutes    int32 `bson:"lockout_minutes"`
}

func toMPasswordPolicy(p *db_models.PasswordPolicy) mPasswordPolicy {
	return mPasswordPolicy{
		ID: p.ID, MinLength: p.MinLength, MaxAgeDays: p.MaxAgeDays,
		RequireUppercase: p.RequireUppercase, RequireLowercase: p.RequireLowercase,
		RequireDigit: p.RequireDigit, RequireSpecial: p.RequireSpecial,
		HistoryCount: p.HistoryCount, MaxLoginFailure: p.MaxLoginFailure,
		LockoutMinutes: p.LockoutMinutes,
	}
}

func fromMPasswordPolicy(m *mPasswordPolicy) *db_models.PasswordPolicy {
	return &db_models.PasswordPolicy{
		ID: m.ID, MinLength: m.MinLength, MaxAgeDays: m.MaxAgeDays,
		RequireUppercase: m.RequireUppercase, RequireLowercase: m.RequireLowercase,
		RequireDigit: m.RequireDigit, RequireSpecial: m.RequireSpecial,
		HistoryCount: m.HistoryCount, MaxLoginFailure: m.MaxLoginFailure,
		LockoutMinutes: m.LockoutMinutes,
	}
}

type mPasswordHistory struct {
	ID           int64     `bson:"id"`
	UserID       int64     `bson:"user_id"`
	PasswordHash string    `bson:"password_hash"`
	ChangedAt    time.Time `bson:"changed_at"`
}

type mUserAccessList struct {
	ID        int64     `bson:"id"`
	ListType  string    `bson:"list_type"`
	MatchType string    `bson:"match_type"`
	Pattern   string    `bson:"pattern"`
	Reason    string    `bson:"reason"`
	CreatedAt time.Time `bson:"created_at"`
}

type mOperationHistory struct {
	ID           int32     `bson:"id"`
	Account      string    `bson:"account"`
	CmdText      string    `bson:"cmd_text"`
	NeNamespace  string    `bson:"ne_namespace"`
	NeIP         string    `bson:"ne_ip"`
	IPAddress    string    `bson:"ip_address"`
	Scope        string    `bson:"scope"`
	Result       string    `bson:"result"`
	CreatedDate  time.Time `bson:"created_date"`
	ExecutedTime time.Time `bson:"executed_time"`
}

func toMOperationHistory(h db_models.OperationHistory) mOperationHistory {
	return mOperationHistory{
		ID: h.ID, Account: h.Account, CmdText: h.CmdText,
		NeNamespace: h.NeNamespace, NeIP: h.NeIP, IPAddress: h.IPAddress,
		Scope: h.Scope, Result: h.Result,
		CreatedDate: h.CreatedDate, ExecutedTime: h.ExecutedTime,
	}
}

func fromMOperationHistory(m *mOperationHistory) db_models.OperationHistory {
	return db_models.OperationHistory{
		ID: m.ID, Account: m.Account, CmdText: m.CmdText,
		NeNamespace: m.NeNamespace, NeIP: m.NeIP, IPAddress: m.IPAddress,
		Scope: m.Scope, Result: m.Result,
		CreatedDate: m.CreatedDate, ExecutedTime: m.ExecutedTime,
	}
}

type mLoginHistory struct {
	ID        int32     `bson:"id"`
	Username  string    `bson:"username"`
	IPAddress string    `bson:"ip_address"`
	TimeLogin time.Time `bson:"time_login"`
}

type mConfigBackup struct {
	ID        int64     `bson:"id"`
	NeName    string    `bson:"ne_name"`
	NeIP      string    `bson:"ne_ip"`
	FilePath  string    `bson:"file_path"`
	Size      int64     `bson:"size"`
	CreatedAt time.Time `bson:"created_at"`
}

func toMConfigBackup(b *db_models.ConfigBackup) mConfigBackup {
	return mConfigBackup{
		ID: b.ID, NeName: b.NeName, NeIP: b.NeIP,
		FilePath: b.FilePath, Size: b.Size, CreatedAt: b.CreatedAt,
	}
}

func fromMConfigBackup(m *mConfigBackup) *db_models.ConfigBackup {
	return &db_models.ConfigBackup{
		ID: m.ID, NeName: m.NeName, NeIP: m.NeIP,
		FilePath: m.FilePath, Size: m.Size, CreatedAt: m.CreatedAt,
	}
}
