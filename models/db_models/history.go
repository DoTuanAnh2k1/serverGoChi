package db_models

import "time"

// OperationHistory is the audit trail. Survives user purges intentionally —
// even after a user row is deleted, their actions remain searchable by
// their (then-current) username. Scope = "ne-config" | "ne-command" | "mgt".
const TableNameOperationHistory = "operation_history"

type OperationHistory struct {
	ID           int32     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Account      string    `gorm:"column:account;type:varchar(64);index" json:"account"`
	CmdText      string    `gorm:"column:cmd_text;type:varchar(512)" json:"cmd_text"`
	NeNamespace  string    `gorm:"column:ne_namespace;type:varchar(64)" json:"ne_namespace"`
	NeIP         string    `gorm:"column:ne_ip;type:varchar(64)" json:"ne_ip"`
	IPAddress    string    `gorm:"column:ip_address;type:varchar(64)" json:"ip_address"`
	Scope        string    `gorm:"column:scope;type:varchar(16);index" json:"scope"`
	Result       string    `gorm:"column:result;type:varchar(255)" json:"result"`
	CreatedDate  time.Time `gorm:"column:created_date;index;autoCreateTime" json:"created_date"`
	ExecutedTime time.Time `gorm:"column:executed_time" json:"executed_time"`
}

func (*OperationHistory) TableName() string { return TableNameOperationHistory }

// LoginHistory tracks every successful login (after auth gates).
const TableNameLoginHistory = "login_history"

type LoginHistory struct {
	ID        int32     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"column:username;type:varchar(64);index" json:"username"`
	IPAddress string    `gorm:"column:ip_address;type:varchar(64)" json:"ip_address"`
	TimeLogin time.Time `gorm:"column:time_login;autoCreateTime" json:"time_login"`
}

func (*LoginHistory) TableName() string { return TableNameLoginHistory }

// ConfigBackup is preserved verbatim from v1 — orthogonal to RBAC.
const TableNameConfigBackup = "config_backup"

type ConfigBackup struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	NeName    string    `gorm:"column:ne_name;type:varchar(64);index" json:"ne_name"`
	NeIP      string    `gorm:"column:ne_ip;type:varchar(64)" json:"ne_ip"`
	FilePath  string    `gorm:"column:file_path;type:varchar(255)" json:"file_path"`
	Size      int64     `gorm:"column:size" json:"size"`
	CreatedAt time.Time `gorm:"column:created_at;index;autoCreateTime" json:"created_at"`
}

func (*ConfigBackup) TableName() string { return TableNameConfigBackup }
