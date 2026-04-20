package store

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/config"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/repository/mongodb"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/repository/mysql"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/repository/postgres"
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
	case "mysql", "mariadb":
		store = mysql.GetInstance()
	case "mongodb":
		store = mongodb.GetInstance()
	case "postgres":
		store = postgres.GetInstance()
	default:
		panic("unsupported database type: " + cfg.DbType)
	}

	err := store.Init(cfg)
	if err != nil {
		panic("store: failed to connect: " + err.Error())
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
	GetCliNeListBySystemType(string) ([]*db_models.CliNe, error)
	GetCliNeByNeId(int64) (*db_models.CliNe, error)
	CreateCliNe(*db_models.CliNe) error
	UpdateCliNe(*db_models.CliNe) error
	DeleteCliNeById(int64) error
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

	// GetRecentHistoryFiltered returns N most recent records filtered by scope, NE name, and/or account.
	GetRecentHistoryFiltered(limit int, scope, neName, account string) ([]db_models.CliOperationHistory, error)

	// Leader-only: get all history for a given day for CSV export
	GetDailyOperationHistory(date time.Time) ([]db_models.CliOperationHistory, error)

	// Leader-only: delete history before cutoff, returns deleted count.
	DeleteHistoryBefore(cutoff time.Time) (int64, error)

	// config-backup — NETCONF commit backup (metadata in DB, XML on disk)
	SaveConfigBackup(b *db_models.CliConfigBackup) error
	ListConfigBackups(neName string) ([]*db_models.CliConfigBackup, error)
	GetConfigBackupById(id int64) (*db_models.CliConfigBackup, error)

	// cli_group — CRUD
	CreateGroup(g *db_models.CliGroup) error
	GetGroupById(id int64) (*db_models.CliGroup, error)
	GetGroupByName(name string) (*db_models.CliGroup, error)
	GetAllGroups() ([]*db_models.CliGroup, error)
	UpdateGroup(g *db_models.CliGroup) error
	DeleteGroupById(id int64) error

	// cli_user_group_mapping
	CreateUserGroupMapping(m *db_models.CliUserGroupMapping) error
	DeleteUserGroupMapping(m *db_models.CliUserGroupMapping) error
	GetAllGroupsOfUser(userId int64) ([]*db_models.CliUserGroupMapping, error)
	GetAllUsersOfGroup(groupId int64) ([]*db_models.CliUserGroupMapping, error)
	DeleteAllUserGroupMappingByUserId(userId int64) error
	DeleteAllUserGroupMappingByGroupId(groupId int64) error

	// cli_group_ne_mapping
	CreateGroupNeMapping(m *db_models.CliGroupNeMapping) error
	DeleteGroupNeMapping(m *db_models.CliGroupNeMapping) error
	GetAllNesOfGroup(groupId int64) ([]*db_models.CliGroupNeMapping, error)
	GetAllGroupsOfNe(neId int64) ([]*db_models.CliGroupNeMapping, error)
	DeleteAllGroupNeMappingByGroupId(groupId int64) error
	DeleteAllGroupNeMappingByNeId(neId int64) error
}
