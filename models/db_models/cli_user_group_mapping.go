package db_models

const TableNameCliUserGroupMapping = "cli_user_group_mapping"

// CliUserGroupMapping mapped from table <cli_user_group_mapping>
type CliUserGroupMapping struct {
	UserID  int64 `gorm:"column:user_id;primaryKey" json:"user_id"`
	GroupID int64 `gorm:"column:group_id;primaryKey" json:"group_id"`
}

func (*CliUserGroupMapping) TableName() string { return TableNameCliUserGroupMapping }
