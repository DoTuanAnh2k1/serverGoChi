package db_models

const TableNameCliNe = "cli_ne"

// CliNe mapped from table <cli_ne>
type CliNe struct {
	ID                int64  `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	NeName            string `gorm:"column:ne_name" json:"ne_name"`
	Namespace         string `gorm:"column:namespace" json:"namespace"`
	SiteName          string `gorm:"column:site_name" json:"site_name"`
	SystemType        string `gorm:"column:system_type" json:"system_type"`
	Description       string `gorm:"column:description" json:"description"`
	CommandURL        string `gorm:"column:command_url" json:"command_url"`
	ConfMode          string `gorm:"column:conf_mode" json:"conf_mode"`
	ConfMasterIP      string `gorm:"column:conf_master_ip" json:"conf_master_ip"`
	ConfSlaveIP       string `gorm:"column:conf_slave_ip" json:"conf_slave_ip"`
	ConfPortMasterSSH int32  `gorm:"column:conf_port_master_ssh" json:"conf_port_master_ssh"`
	ConfPortSlaveSSH  int32  `gorm:"column:conf_port_slave_ssh" json:"conf_port_slave_ssh"`
	ConfPortMasterTCP int32  `gorm:"column:conf_port_master_tcp" json:"conf_port_master_tcp"`
	ConfPortSlaveTCP  int32  `gorm:"column:conf_port_slave_tcp" json:"conf_port_slave_tcp"`
	ConfUsername      string `gorm:"column:conf_username" json:"conf_username"`
	ConfPassword      string `gorm:"column:conf_password" json:"conf_password"`
}

// TableName CliNe's table name
func (*CliNe) TableName() string {
	return TableNameCliNe
}
