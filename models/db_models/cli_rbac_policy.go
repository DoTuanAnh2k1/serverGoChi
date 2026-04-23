package db_models

import "time"

// Password policy + history + management permissions — the remaining RBAC
// design items (docs/rbac-design.md §4.8, §4.11, §5.2) implemented on top of
// the core RBAC in cli_rbac.go.

// ────────── cli_password_policy ──────────

const TableNameCliPasswordPolicy = "cli_password_policy"

// CliPasswordPolicy is attached to a group via cli_group.password_policy_id.
// When a user belongs to multiple groups, the effective policy is the
// strict-est value for each field — see service.GetEffectivePasswordPolicy.
type CliPasswordPolicy struct {
	ID               int64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name             string `gorm:"column:name;type:varchar(64);uniqueIndex" json:"name"`
	MaxAgeDays       int32  `gorm:"column:max_age_days;default:0" json:"max_age_days"`
	MinLength        int32  `gorm:"column:min_length;default:8" json:"min_length"`
	RequireUppercase bool   `gorm:"column:require_uppercase" json:"require_uppercase"`
	RequireLowercase bool   `gorm:"column:require_lowercase" json:"require_lowercase"`
	RequireDigit     bool   `gorm:"column:require_digit" json:"require_digit"`
	RequireSpecial   bool   `gorm:"column:require_special" json:"require_special"`
	HistoryCount     int32  `gorm:"column:history_count;default:0" json:"history_count"`
	MaxLoginFailure  int32  `gorm:"column:max_login_failure;default:0" json:"max_login_failure"`
	LockoutMinutes   int32  `gorm:"column:lockout_minutes;default:0" json:"lockout_minutes"`
}

func (*CliPasswordPolicy) TableName() string { return TableNameCliPasswordPolicy }

// ────────── cli_password_history ──────────

const TableNameCliPasswordHistory = "cli_password_history"

// CliPasswordHistory stores past password hashes so the policy can enforce
// "no reuse of last N passwords". Entries older than the user's effective
// history_count are pruned on each change.
type CliPasswordHistory struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID       int64     `gorm:"column:user_id;index" json:"user_id"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(256)" json:"password_hash"`
	ChangedAt    time.Time `gorm:"column:changed_at" json:"changed_at"`
}

func (*CliPasswordHistory) TableName() string { return TableNameCliPasswordHistory }

// ────────── cli_group_mgt_permission ──────────

const TableNameCliGroupMgtPermission = "cli_group_mgt_permission"

// CliGroupMgtPermission grants a group access to a (resource, action) pair
// on the mgt-svc itself. Companion to CliGroupCmdPermission — the former
// is about NE/command authorization, this is about who can manage the
// management plane (users, groups, NEs, etc.). Supports "*" wildcards.
type CliGroupMgtPermission struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupID  int64  `gorm:"column:group_id;index" json:"group_id"`
	Resource string `gorm:"column:resource;type:varchar(32)" json:"resource"`
	Action   string `gorm:"column:action;type:varchar(16)" json:"action"`
}

func (*CliGroupMgtPermission) TableName() string { return TableNameCliGroupMgtPermission }

// Standard resources & actions for the mgt-svc. Handlers can use the "*"
// wildcard for either side to represent "all".
const (
	MgtResourceAny      = "*"
	MgtResourceUser     = "user"
	MgtResourceNe       = "ne"
	MgtResourceGroup    = "group"
	MgtResourceCommand  = "command"
	MgtResourcePolicy   = "policy"
	MgtResourceHistory  = "history"

	MgtActionAny    = "*"
	MgtActionCreate = "create"
	MgtActionRead   = "read"
	MgtActionUpdate = "update"
	MgtActionDelete = "delete"
)
