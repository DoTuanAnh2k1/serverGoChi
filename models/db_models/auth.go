package db_models

import "time"

// Auth-side controls: a single global password policy, password history (for
// the no-reuse-last-N rule) and a flat blacklist/whitelist of access list
// entries that gate authentication itself.

const TableNamePasswordPolicy = "password_policy"

// PasswordPolicy is a singleton row (id always 1). v1 had per-group policies;
// v2 collapses to one global policy because the new model has no role.
// max_login_failure / lockout_minutes are 0 to disable lockout entirely.
type PasswordPolicy struct {
	ID                int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MinLength         int32 `gorm:"column:min_length;default:8" json:"min_length"`
	MaxAgeDays        int32 `gorm:"column:max_age_days;default:0" json:"max_age_days"`
	RequireUppercase  bool  `gorm:"column:require_uppercase;default:false" json:"require_uppercase"`
	RequireLowercase  bool  `gorm:"column:require_lowercase;default:false" json:"require_lowercase"`
	RequireDigit      bool  `gorm:"column:require_digit;default:false" json:"require_digit"`
	RequireSpecial    bool  `gorm:"column:require_special;default:false" json:"require_special"`
	HistoryCount      int32 `gorm:"column:history_count;default:0" json:"history_count"`
	MaxLoginFailure   int32 `gorm:"column:max_login_failure;default:0" json:"max_login_failure"`
	LockoutMinutes    int32 `gorm:"column:lockout_minutes;default:0" json:"lockout_minutes"`
}

func (*PasswordPolicy) TableName() string { return TableNamePasswordPolicy }

const TableNamePasswordHistory = "password_history"

type PasswordHistory struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID       int64     `gorm:"column:user_id;index" json:"user_id"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(256)" json:"-"`
	ChangedAt    time.Time `gorm:"column:changed_at;autoCreateTime" json:"changed_at"`
}

func (*PasswordHistory) TableName() string { return TableNamePasswordHistory }

// ── access list ──

const TableNameUserAccessList = "user_access_list"

const (
	AccessListTypeBlacklist = "blacklist"
	AccessListTypeWhitelist = "whitelist"

	AccessListMatchUsername    = "username"     // exact match on User.Username
	AccessListMatchIPCidr      = "ip_cidr"      // CIDR match on remote IP at login
	AccessListMatchEmailDomain = "email_domain" // suffix match on email
)

// UserAccessList gates authentication itself. Rules:
//   - any blacklist entry that matches → DENY login
//   - if any whitelist entry exists for the same match_type, the user must
//     match at least one of them → otherwise DENY login
//   - whitelist with zero entries means "allow all" (no restriction)
//
// (list_type, match_type, pattern) is unique so the same rule can't be
// inserted twice.
type UserAccessList struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ListType  string    `gorm:"column:list_type;type:varchar(16);index;uniqueIndex:uq_acl_sig" json:"list_type"`
	MatchType string    `gorm:"column:match_type;type:varchar(16);uniqueIndex:uq_acl_sig" json:"match_type"`
	Pattern   string    `gorm:"column:pattern;type:varchar(255);uniqueIndex:uq_acl_sig" json:"pattern"`
	Reason    string    `gorm:"column:reason;type:varchar(255)" json:"reason"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (*UserAccessList) TableName() string { return TableNameUserAccessList }
