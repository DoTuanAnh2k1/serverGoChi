package db_models

const TableNameCliNeConfig = "cli_ne_config"

// CliNeConfig stores IP/connection config for a NE in ne-config mode.
type CliNeConfig struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	NeID        int64  `gorm:"column:ne_id;not null" json:"ne_id"`
	IPAddress   string `gorm:"column:ip_address;not null" json:"ip_address"`
	Port        int32  `gorm:"column:port" json:"port"`
	Username    string `gorm:"column:username" json:"username"`
	Password    string `gorm:"column:password" json:"password"`
	Protocol    string `gorm:"column:protocol;default:SSH" json:"protocol"`
	Description string `gorm:"column:description" json:"description"`
}

func (*CliNeConfig) TableName() string {
	return TableNameCliNeConfig
}
