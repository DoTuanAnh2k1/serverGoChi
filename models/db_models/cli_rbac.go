package db_models

// This file defines the RBAC entities introduced by docs/rbac-design.md.
//
// Model:
//   - CliNeProfile classifies an NE by command set (SMF, AMF, UPF, generic-router).
//   - CliCommandDef is a single command pattern that is defined per ne_profile.
//   - CliCommandGroup groups related commands (all share the same ne_profile).
//   - CliCommandGroupMapping is the M:N bridge between groups and defs.
//   - CliGroupCmdPermission grants (or denies) a group a command/group/category
//     at a given ne_scope ("*" | "profile:<p>" | "ne:<name>").
//
// The evaluation algorithm (see service/rbac/evaluator.go) combines AWS-IAM
// "explicit deny > explicit allow > implicit deny" with Vault-style
// scope-specificity (ne:X beats profile:Y beats *).

// ────────── cli_ne_profile ──────────

const TableNameCliNeProfile = "cli_ne_profile"

// CliNeProfile is a functional classification for an NE.
type CliNeProfile struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;type:varchar(64);uniqueIndex" json:"name"`
	Description string `gorm:"column:description;type:varchar(512)" json:"description"`
}

func (*CliNeProfile) TableName() string { return TableNameCliNeProfile }

// ────────── cli_command_def ──────────

const TableNameCliCommandDef = "cli_command_def"

// Command services supported by the registry.
const (
	CommandServiceNeCommand = "ne-command"
	CommandServiceNeConfig  = "ne-config"
	CommandServiceAny       = "*"
)

// Command categories — informational, also usable as a grant_type value.
const (
	CommandCategoryMonitoring    = "monitoring"
	CommandCategoryConfiguration = "configuration"
	CommandCategoryAdmin         = "admin"
	CommandCategoryDebug         = "debug"
)

// CliCommandDef is one command pattern, scoped to a single ne_profile (or "*"
// for commands valid on any NE).
type CliCommandDef struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Service     string `gorm:"column:service;type:varchar(32);index" json:"service"`
	NeProfile   string `gorm:"column:ne_profile;type:varchar(64);index" json:"ne_profile"`
	Pattern     string `gorm:"column:pattern;type:varchar(256)" json:"pattern"`
	Category    string `gorm:"column:category;type:varchar(32);index" json:"category"`
	RiskLevel   int32  `gorm:"column:risk_level" json:"risk_level"`
	Description string `gorm:"column:description;type:varchar(512)" json:"description"`
	CreatedBy   string `gorm:"column:created_by;type:varchar(64)" json:"created_by"`
}

func (*CliCommandDef) TableName() string { return TableNameCliCommandDef }

// ────────── cli_command_group ──────────

const TableNameCliCommandGroup = "cli_command_group"

// CliCommandGroup bundles related commands of the same profile, so that a
// permission rule can reference a whole bundle instead of listing patterns.
type CliCommandGroup struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;type:varchar(64);uniqueIndex" json:"name"`
	NeProfile   string `gorm:"column:ne_profile;type:varchar(64)" json:"ne_profile"`
	Service     string `gorm:"column:service;type:varchar(32)" json:"service"`
	Description string `gorm:"column:description;type:varchar(512)" json:"description"`
	CreatedBy   string `gorm:"column:created_by;type:varchar(64)" json:"created_by"`
}

func (*CliCommandGroup) TableName() string { return TableNameCliCommandGroup }

// ────────── cli_command_group_mapping ──────────

const TableNameCliCommandGroupMapping = "cli_command_group_mapping"

// CliCommandGroupMapping is the M:N bridge between command groups and defs.
type CliCommandGroupMapping struct {
	CommandGroupID int64 `gorm:"column:command_group_id;primaryKey" json:"command_group_id"`
	CommandDefID   int64 `gorm:"column:command_def_id;primaryKey" json:"command_def_id"`
}

func (*CliCommandGroupMapping) TableName() string { return TableNameCliCommandGroupMapping }

// ────────── cli_group_cmd_permission ──────────

const TableNameCliGroupCmdPermission = "cli_group_cmd_permission"

// Grant types — what the permission references.
const (
	GrantTypeCommandGroup = "command_group"
	GrantTypeCategory     = "category"
	GrantTypePattern      = "pattern"
)

// Effect values — AWS-IAM style.
const (
	PermissionEffectAllow = "allow"
	PermissionEffectDeny  = "deny"
)

// Common ne_scope values. Actual scope strings take one of the forms
//   - "*"              global
//   - "profile:<name>" all NEs with a matching profile
//   - "ne:<ne_name>"   a specific NE
const (
	NeScopeAny            = "*"
	NeScopePrefixProfile  = "profile:"
	NeScopePrefixSpecific = "ne:"
)

// CliGroupCmdPermission grants/denies a group the ability to run a command
// (pattern / category / command_group) within a given NE scope.
type CliGroupCmdPermission struct {
	ID         int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	GroupID    int64  `gorm:"column:group_id;index" json:"group_id"`
	Service    string `gorm:"column:service;type:varchar(32)" json:"service"`
	NeScope    string `gorm:"column:ne_scope;type:varchar(128)" json:"ne_scope"`
	GrantType  string `gorm:"column:grant_type;type:varchar(16)" json:"grant_type"`
	GrantValue string `gorm:"column:grant_value;type:varchar(256)" json:"grant_value"`
	Effect     string `gorm:"column:effect;type:varchar(8)" json:"effect"`
}

func (*CliGroupCmdPermission) TableName() string { return TableNameCliGroupCmdPermission }
