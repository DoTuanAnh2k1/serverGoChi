package db_models

const TableNameCliGroup = "cli_group"

// CliGroup mapped from table <cli_group>
type CliGroup struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;type:varchar(255);uniqueIndex" json:"name"`
	Description string `gorm:"column:description;type:varchar(255)" json:"description"`
	// PasswordPolicyID is the per-group password policy (optional — nil means
	// no constraints beyond whatever the change-password handler enforces
	// globally).
	PasswordPolicyID *int64 `gorm:"column:password_policy_id" json:"password_policy_id,omitempty"`
}

func (*CliGroup) TableName() string { return TableNameCliGroup }
