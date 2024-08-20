package store

import (
	"serverGoChi/models/config_models"
	"serverGoChi/models/db_models"
	"serverGoChi/src/db/mysql"
	"time"
)

var (
	store DatabaseStore
)

func GetSingleton() DatabaseStore {
	return store
}

func Init(cfg *config_models.Config) {

	switch cfg.Db.DbType {
	case "mysql":
		store = mysql.GetInstance()
	case "mongodb":
		// store = mongodb.GetInstance()
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
	Init(cfg *config_models.Config) error

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

	GetCliRole(db_models.CliRole) (*db_models.CliRole, error)
	CreateCliRole(db_models.CliRole) error
	DeleteCliRole(db_models.CliRole) error
	GetAllCliRole() ([]*db_models.CliRole, error)
	GetCliNeListBySystemType(string) ([]*db_models.CliNe, error)
	GetCliNeByNeId(int64) (*db_models.CliNe, error)
}
