package db_models

// v2 has TWO group concepts. Each glues a set of users to a set of resources:
//
//   ne_access_group  →  who can REACH which NEs
//   cmd_exec_group   →  who can EXECUTE which commands
//
// They are independent — a user can be in many of either kind, and a single
// rule "user X may run command Y on NE Z" requires BOTH layers to allow it
// (NE access ∋ Z and command exec ∋ Y; Y itself already names which NE it
// belongs to).

// ── NE Access Group ──

const (
	TableNameNeAccessGroup     = "ne_access_group"
	TableNameNeAccessGroupUser = "ne_access_group_user"
	TableNameNeAccessGroupNe   = "ne_access_group_ne"
)

type NeAccessGroup struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;type:varchar(64);uniqueIndex" json:"name"`
	Description string `gorm:"column:description;type:varchar(255)" json:"description"`
}

func (*NeAccessGroup) TableName() string { return TableNameNeAccessGroup }

type NeAccessGroupUser struct {
	GroupID int64 `gorm:"column:group_id;primaryKey" json:"group_id"`
	UserID  int64 `gorm:"column:user_id;primaryKey" json:"user_id"`
}

func (*NeAccessGroupUser) TableName() string { return TableNameNeAccessGroupUser }

type NeAccessGroupNe struct {
	GroupID int64 `gorm:"column:group_id;primaryKey" json:"group_id"`
	NeID    int64 `gorm:"column:ne_id;primaryKey" json:"ne_id"`
}

func (*NeAccessGroupNe) TableName() string { return TableNameNeAccessGroupNe }

// ── Command Exec Group ──

const (
	TableNameCmdExecGroup        = "cmd_exec_group"
	TableNameCmdExecGroupUser    = "cmd_exec_group_user"
	TableNameCmdExecGroupCommand = "cmd_exec_group_command"
)

type CmdExecGroup struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"column:name;type:varchar(64);uniqueIndex" json:"name"`
	Description string `gorm:"column:description;type:varchar(255)" json:"description"`
}

func (*CmdExecGroup) TableName() string { return TableNameCmdExecGroup }

type CmdExecGroupUser struct {
	GroupID int64 `gorm:"column:group_id;primaryKey" json:"group_id"`
	UserID  int64 `gorm:"column:user_id;primaryKey" json:"user_id"`
}

func (*CmdExecGroupUser) TableName() string { return TableNameCmdExecGroupUser }

type CmdExecGroupCommand struct {
	GroupID   int64 `gorm:"column:group_id;primaryKey" json:"group_id"`
	CommandID int64 `gorm:"column:command_id;primaryKey" json:"command_id"`
}

func (*CmdExecGroupCommand) TableName() string { return TableNameCmdExecGroupCommand }
