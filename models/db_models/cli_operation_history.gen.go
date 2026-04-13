package db_models

import "time"

const TableNameCliOperationHistory = "cli_operation_history"

// CliOperationHistory mapped from table <cli_operation_history>
type CliOperationHistory struct {
	ID           int32     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Account      string    `gorm:"column:account;not null" json:"account"`
	CmdName      string    `gorm:"column:cmd_name;not null" json:"cmd_name"`
	NeName       string    `gorm:"column:ne_name" json:"ne_name"`
	NeIP         string    `gorm:"column:ne_ip" json:"ne_ip"`
	IPAddress    string    `gorm:"column:ip_address" json:"ip_address"`
	Scope        string    `gorm:"column:scope" json:"scope"`
	Result       string    `gorm:"column:result" json:"result"`
	CreatedDate  time.Time `gorm:"column:created_date;not null" json:"created_date"`
	ExecutedTime time.Time `gorm:"column:executed_time" json:"executed_time"`
}

// TableName CliOperationHistory's table name
func (*CliOperationHistory) TableName() string {
	return TableNameCliOperationHistory
}
