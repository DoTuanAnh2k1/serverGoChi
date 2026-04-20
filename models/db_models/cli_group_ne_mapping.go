package db_models

const TableNameCliGroupNeMapping = "cli_group_ne_mapping"

// CliGroupNeMapping mapped from table <cli_group_ne_mapping>
type CliGroupNeMapping struct {
	GroupID int64 `gorm:"column:group_id;primaryKey" json:"group_id"`
	TblNeID int64 `gorm:"column:tbl_ne_id;primaryKey" json:"tbl_ne_id"`
}

func (*CliGroupNeMapping) TableName() string { return TableNameCliGroupNeMapping }
