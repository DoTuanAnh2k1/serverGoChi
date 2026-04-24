package db_models

const TableNameCommand = "command"

// Command service tags.
const (
	CommandServiceNeConfig  = "ne-config"
	CommandServiceNeCommand = "ne-command"
)

// Command is an exact command text registered against one NE. v2 deliberately
// drops pattern matching — registration is verbose but unambiguous, and the
// permission check becomes a simple equality lookup. (ne_id, service,
// cmd_text) is unique; GORM's composite uniqueIndex tag (same name on all
// three fields) creates the constraint during AutoMigrate.
type Command struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	NeID        int64  `gorm:"column:ne_id;index;uniqueIndex:uq_command" json:"ne_id"`
	Service     string `gorm:"column:service;type:varchar(16);index;uniqueIndex:uq_command" json:"service"`
	CmdText     string `gorm:"column:cmd_text;type:varchar(512);uniqueIndex:uq_command" json:"cmd_text"`
	Description string `gorm:"column:description;type:varchar(512)" json:"description"`
}

func (*Command) TableName() string { return TableNameCommand }
