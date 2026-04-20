package db_models

const TableNameCliGroup = "cli_group"

// CliGroup mapped from table <cli_group>
type CliGroup struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;type:varchar(255);uniqueIndex" json:"name"`
	Description string `gorm:"column:description;type:varchar(255)" json:"description"`
}

func (*CliGroup) TableName() string { return TableNameCliGroup }
