// Package store is the v2 storage abstraction. Three drivers (mysql,
// postgres, mongodb) all implement DatabaseStore. Compile-time check is
// the type assertion at the end of Init() — if a driver drifts, the build
// breaks immediately.
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

var store DatabaseStore

func GetSingleton() DatabaseStore { return store }

// SetSingleton lets tests inject a mock.
func SetSingleton(s DatabaseStore) { store = s }

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
	if err := store.Init(cfg); err != nil {
		panic("store: failed to connect: " + err.Error())
	}
}

// DatabaseStore is the v2 surface — flat, intentionally small. Naming
// follows entity-then-action: GetUserByX, ListNEs, AssignUserToNeAccessGroup.
type DatabaseStore interface {
	Init(cfg config_models.DatabaseConfig) error
	Ping() error

	// ── User ──
	CreateUser(u *db_models.User) error
	GetUserByID(id int64) (*db_models.User, error)
	GetUserByUsername(username string) (*db_models.User, error)
	ListUsers() ([]*db_models.User, error)
	UpdateUser(u *db_models.User) error
	DeleteUserByID(id int64) error

	// ── NE ──
	CreateNE(n *db_models.NE) error
	GetNEByID(id int64) (*db_models.NE, error)
	GetNEByNamespace(ns string) (*db_models.NE, error)
	ListNEs() ([]*db_models.NE, error)
	UpdateNE(n *db_models.NE) error
	DeleteNEByID(id int64) error

	// ── Command ──
	CreateCommand(c *db_models.Command) error
	GetCommandByID(id int64) (*db_models.Command, error)
	GetCommandByTriple(neID int64, service, cmdText string) (*db_models.Command, error)
	ListCommands(neID int64, service string) ([]*db_models.Command, error)
	UpdateCommand(c *db_models.Command) error
	DeleteCommandByID(id int64) error

	// ── NE Access Group ──
	CreateNeAccessGroup(g *db_models.NeAccessGroup) error
	GetNeAccessGroupByID(id int64) (*db_models.NeAccessGroup, error)
	GetNeAccessGroupByName(name string) (*db_models.NeAccessGroup, error)
	ListNeAccessGroups() ([]*db_models.NeAccessGroup, error)
	UpdateNeAccessGroup(g *db_models.NeAccessGroup) error
	DeleteNeAccessGroupByID(id int64) error
	AddUserToNeAccessGroup(groupID, userID int64) error
	RemoveUserFromNeAccessGroup(groupID, userID int64) error
	ListUsersInNeAccessGroup(groupID int64) ([]int64, error)
	ListNeAccessGroupsOfUser(userID int64) ([]int64, error)
	AddNeToNeAccessGroup(groupID, neID int64) error
	RemoveNeFromNeAccessGroup(groupID, neID int64) error
	ListNEsInNeAccessGroup(groupID int64) ([]int64, error)
	ListNeAccessGroupsOfNE(neID int64) ([]int64, error)

	// ── Cmd Exec Group ──
	CreateCmdExecGroup(g *db_models.CmdExecGroup) error
	GetCmdExecGroupByID(id int64) (*db_models.CmdExecGroup, error)
	GetCmdExecGroupByName(name string) (*db_models.CmdExecGroup, error)
	ListCmdExecGroups() ([]*db_models.CmdExecGroup, error)
	UpdateCmdExecGroup(g *db_models.CmdExecGroup) error
	DeleteCmdExecGroupByID(id int64) error
	AddUserToCmdExecGroup(groupID, userID int64) error
	RemoveUserFromCmdExecGroup(groupID, userID int64) error
	ListUsersInCmdExecGroup(groupID int64) ([]int64, error)
	ListCmdExecGroupsOfUser(userID int64) ([]int64, error)
	AddCommandToCmdExecGroup(groupID, commandID int64) error
	RemoveCommandFromCmdExecGroup(groupID, commandID int64) error
	ListCommandsInCmdExecGroup(groupID int64) ([]int64, error)
	ListCmdExecGroupsOfCommand(commandID int64) ([]int64, error)

	// ── Password Policy + History ──
	GetPasswordPolicy() (*db_models.PasswordPolicy, error)
	UpsertPasswordPolicy(p *db_models.PasswordPolicy) error
	AppendPasswordHistory(h *db_models.PasswordHistory) error
	GetRecentPasswordHistory(userID int64, limit int) ([]*db_models.PasswordHistory, error)
	PrunePasswordHistory(userID int64, keep int) error

	// ── User Access List ──
	CreateAccessListEntry(e *db_models.UserAccessList) error
	ListAccessListEntries(listType string) ([]*db_models.UserAccessList, error)
	DeleteAccessListEntryByID(id int64) error

	// ── History ──
	SaveOperationHistory(h db_models.OperationHistory) error
	GetRecentHistory(limit int) ([]db_models.OperationHistory, error)
	GetRecentHistoryFiltered(limit int, scope, neNamespace, account string) ([]db_models.OperationHistory, error)
	GetDailyOperationHistory(date time.Time) ([]db_models.OperationHistory, error)
	DeleteHistoryBefore(cutoff time.Time) (int64, error)
	UpdateLoginHistory(username, ip string, t time.Time) error

	// ── Config Backup (orthogonal, kept as-is) ──
	SaveConfigBackup(b *db_models.ConfigBackup) error
	ListConfigBackups(neName string) ([]*db_models.ConfigBackup, error)
	GetConfigBackupByID(id int64) (*db_models.ConfigBackup, error)
}
