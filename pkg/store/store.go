package store

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/config"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/repository/mongodb"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/repository/mysql"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/repository/postgres"
	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

var (
	store DatabaseStore
)

func GetSingleton() DatabaseStore {
	return store
}

// SetSingleton is used in tests to inject a mock store.
func SetSingleton(s DatabaseStore) {
	store = s
}

func Init() {
	cfg := config.GetDatabaseConfig()
	switch cfg.DbType {
	case "mysql":
		store = mysql.GetInstance()
	case "mongodb":
		store = mongodb.GetInstance()
	case "postgres":
		store = postgres.GetInstance()
	default:
		panic("unsupported database type")
	}
	err := store.Init(cfg)
	if err != nil {
		panic("cant init store")
	}
}

// DatabaseStore Database Store
type DatabaseStore interface {
	Init(cfg config_models.DatabaseConfig) error

	GetAllUser() ([]*db_models.TblAccount, error)
	GetUserByUserName(string) (*db_models.TblAccount, error)
	UpdateUser(account *db_models.TblAccount) error
	AddUser(*db_models.TblAccount) error
	Ping() error
	UpdateLoginHistory(string, string, time.Time) error
	SaveHistoryCommand(db_models.CliOperationHistory) error
	GetCLIUserNeMappingByUserId(int64) (*db_models.CliUserNeMapping, error)
	GetNeListById(int64) ([]*db_models.CliNe, error)
	GetRolesById(int64) ([]*db_models.CliRoleUserMapping, error)
	GetCliRole(*db_models.CliRole) (*db_models.CliRole, error)
	CreateCliRole(*db_models.CliRole) error
	DeleteCliRole(*db_models.CliRole) error
	GetAllCliRole() ([]*db_models.CliRole, error)
	GetCliNeListBySystemType(string) ([]*db_models.CliNe, error)
	GetCliNeByNeId(int64) (*db_models.CliNe, error)
	CreateCliNe(*db_models.CliNe) error
	UpdateCliNe(*db_models.CliNe) error
	DeleteCliNeById(int64) error
	AddRole(*db_models.CliRoleUserMapping) error
	DeleteRole(*db_models.CliRoleUserMapping) error
	CreateUserNeMapping(*db_models.CliUserNeMapping) error
	DeleteUserNeMapping(*db_models.CliUserNeMapping) error
	DeleteAllUserNeMappingByNeId(neId int64) error
	GetNeMonitorById(int64) (*db_models.CliNeMonitor, error)
	DeleteNeMonitorByNeId(neId int64) error
	GetAllNeOfUserByUserId(int64) ([]*db_models.CliUserNeMapping, error)
	DeleteCliNeSlaveByNeId(neId int64) error

	// cli_ne_config — IP/connection config for ne-config mode
	CreateCliNeConfig(*db_models.CliNeConfig) error
	GetCliNeConfigByNeId(neId int64) ([]*db_models.CliNeConfig, error)
	GetCliNeConfigById(id int64) (*db_models.CliNeConfig, error)
	UpdateCliNeConfig(*db_models.CliNeConfig) error
	DeleteCliNeConfigById(id int64) error
	DeleteCliNeConfigByNeId(neId int64) error

	// GetRecentHistory returns the N most recent history records.
	GetRecentHistory(limit int) ([]db_models.CliOperationHistory, error)

	// GetRecentHistoryFiltered returns N most recent records filtered by scope and/or NE name.
	GetRecentHistoryFiltered(limit int, scope, neName string) ([]db_models.CliOperationHistory, error)

	// Leader-only: get all history for a given day for CSV export
	GetDailyOperationHistory(date time.Time) ([]db_models.CliOperationHistory, error)

	// Leader-only: delete history before cutoff, returns deleted count.
	DeleteHistoryBefore(cutoff time.Time) (int64, error)

	// config-backup — NETCONF commit backup (metadata in DB, XML on disk)
	SaveConfigBackup(b *db_models.CliConfigBackup) error
	ListConfigBackups(neName string) ([]*db_models.CliConfigBackup, error)
	GetConfigBackupById(id int64) (*db_models.CliConfigBackup, error)
}
