package db_models

const TableNameCliGroup = "cli_group"

// CliGroup mapped from table <cli_group>
type CliGroup struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;uniqueIndex" json:"name"`
	Description string `gorm:"column:description" json:"description"`
}

func (*CliGroup) TableName() string { return TableNameCliGroup }
