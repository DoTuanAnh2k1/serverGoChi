package db_models

const TableNameNe = "ne"

// NE constants.
const (
	NeConfModeSSH      = "SSH"
	NeConfModeTelnet   = "TELNET"
	NeConfModeNetconf  = "NETCONF"
	NeConfModeRestconf = "RESTCONF"
)

// NE represents a managed network element. The v1 cli_ne had `ne_name` as
// the display label and an open-ended `system_type` — v2 swaps them:
// `Namespace` is the unique identifier (e.g. "htsmf01") and `NeType` is the
// functional category ("SMF", "AMF", "UPF", "PE", ...). Commands live on
// each NE individually (no profile-wide registration) so the model stays
// small.
type NE struct {
	ID           int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Namespace    string `gorm:"column:namespace;type:varchar(64);uniqueIndex" json:"namespace"`
	NeType       string `gorm:"column:ne_type;type:varchar(32);index" json:"ne_type"`
	SiteName     string `gorm:"column:site_name;type:varchar(64)" json:"site_name"`
	Description  string `gorm:"column:description;type:varchar(255)" json:"description"`
	MasterIP     string `gorm:"column:master_ip;type:varchar(64)" json:"master_ip"`
	MasterPort   int32  `gorm:"column:master_port" json:"master_port"`
	SSHUsername  string `gorm:"column:ssh_username;type:varchar(64)" json:"ssh_username"`
	SSHPassword  string `gorm:"column:ssh_password;type:varchar(255)" json:"-"`
	CommandURL   string `gorm:"column:command_url;type:varchar(255)" json:"command_url"`
	ConfMode     string `gorm:"column:conf_mode;type:varchar(32)" json:"conf_mode"`
}

func (*NE) TableName() string { return TableNameNe }
