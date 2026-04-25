// Package db_models contains the v2 schema.
//
// One question to answer: "Can user X execute command Y on NE Z?"
//   1. user X exists, enabled, not locked, not blacklisted, on whitelist (if any)
//   2. user X ∈ some ne_access_group containing NE Z
//   3. user X ∈ some cmd_exec_group containing command Y (where Y is keyed by
//      the (ne_id, service, cmd_text) triple — so Y already implies NE Z)
//
// Three permission tiers: super_admin > admin > user.
//   - super_admin: full access, cannot be deleted by anyone except another super_admin
//   - admin: can configure everything, but cannot delete/demote super_admin accounts
//   - user: read-only on the management UI
package db_models

import "time"

const TableNameUser = "user"

// Role constants.
const (
	RoleSuperAdmin = "super_admin"
	RoleAdmin      = "admin"
	RoleUser       = "user"
)

// RoleLevel returns a numeric weight for comparison (higher = more privileged).
func RoleLevel(role string) int {
	switch role {
	case RoleSuperAdmin:
		return 3
	case RoleAdmin:
		return 2
	default:
		return 1
	}
}

type User struct {
	ID                int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username          string     `gorm:"column:username;type:varchar(64);uniqueIndex" json:"username"`
	PasswordHash      string     `gorm:"column:password_hash;type:varchar(256)" json:"-"`
	Email             string     `gorm:"column:email;type:varchar(128)" json:"email"`
	FullName          string     `gorm:"column:full_name;type:varchar(128)" json:"full_name"`
	Phone             string     `gorm:"column:phone;type:varchar(32)" json:"phone"`
	Role              string     `gorm:"column:role;type:varchar(16);default:user" json:"role"`
	IsEnabled         bool       `gorm:"column:is_enabled;default:true" json:"is_enabled"`
	PasswordExpiresAt *time.Time `gorm:"column:password_expires_at" json:"password_expires_at,omitempty"`
	LoginFailureCount int32      `gorm:"column:login_failure_count;default:0" json:"login_failure_count"`
	LockedAt          *time.Time `gorm:"column:locked_at" json:"locked_at,omitempty"`
	LastLoginAt       *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (*User) TableName() string { return TableNameUser }
