package db_models

import "time"

// CliConfigBackup lưu metadata của một bản backup config NETCONF.
// Nội dung XML được lưu trên disk, DB chỉ lưu metadata và đường dẫn file.
type CliConfigBackup struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	NeName    string    `gorm:"column:ne_name;not null"               json:"ne_name"`
	NeIP      string    `gorm:"column:ne_ip"                          json:"ne_ip"`
	FilePath  string    `gorm:"column:file_path;not null"             json:"-"` // không expose ra API
	Size      int64     `gorm:"column:size"                           json:"size"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"      json:"created_at"`
}

func (*CliConfigBackup) TableName() string { return "cli_config_backup" }
