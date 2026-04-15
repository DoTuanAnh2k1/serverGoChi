// Package testutil provides helpers for unit tests.
// Do not import in production code.
package testutil

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/models/config_models"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// MockStore is a full mock of store.DatabaseStore.
// Each method has a corresponding function field; returns zero value if unset.
type MockStore struct {
	InitFn                        func(cfg config_models.DatabaseConfig) error
	GetAllUserFn                  func() ([]*db_models.TblAccount, error)
	GetUserByUserNameFn           func(name string) (*db_models.TblAccount, error)
	UpdateUserFn                  func(account *db_models.TblAccount) error
	AddUserFn                     func(account *db_models.TblAccount) error
	PingFn                        func() error
	UpdateLoginHistoryFn          func(username, ip string, t time.Time) error
	SaveHistoryCommandFn          func(h db_models.CliOperationHistory) error
	GetCLIUserNeMappingByUserIdFn func(userID int64) (*db_models.CliUserNeMapping, error)
	GetNeListByIdFn               func(id int64) ([]*db_models.CliNe, error)
	GetCliNeListBySystemTypeFn    func(systemType string) ([]*db_models.CliNe, error)
	GetCliNeByNeIdFn              func(id int64) (*db_models.CliNe, error)
	CreateCliNeFn                 func(ne *db_models.CliNe) error
	UpdateCliNeFn                 func(ne *db_models.CliNe) error
	DeleteCliNeByIdFn             func(id int64) error
	CreateUserNeMappingFn            func(m *db_models.CliUserNeMapping) error
	DeleteUserNeMappingFn            func(m *db_models.CliUserNeMapping) error
	DeleteAllUserNeMappingByNeIdFn   func(neId int64) error
	GetNeMonitorByIdFn               func(id int64) (*db_models.CliNeMonitor, error)
	DeleteNeMonitorByNeIdFn          func(neId int64) error
	GetAllNeOfUserByUserIdFn         func(userID int64) ([]*db_models.CliUserNeMapping, error)
	DeleteCliNeSlaveByNeIdFn         func(neId int64) error
	CreateCliNeConfigFn              func(cfg *db_models.CliNeConfig) error
	GetCliNeConfigByNeIdFn           func(neId int64) ([]*db_models.CliNeConfig, error)
	GetCliNeConfigByIdFn             func(id int64) (*db_models.CliNeConfig, error)
	UpdateCliNeConfigFn              func(cfg *db_models.CliNeConfig) error
	DeleteCliNeConfigByIdFn          func(id int64) error
	DeleteCliNeConfigByNeIdFn        func(neId int64) error
	GetRecentHistoryFn               func(limit int) ([]db_models.CliOperationHistory, error)
	GetRecentHistoryFilteredFn       func(limit int, scope, neName string) ([]db_models.CliOperationHistory, error)
	GetDailyOperationHistoryFn       func(date time.Time) ([]db_models.CliOperationHistory, error)
	DeleteHistoryBeforeFn            func(cutoff time.Time) (int64, error)
	SaveConfigBackupFn               func(b *db_models.CliConfigBackup) error
	ListConfigBackupsFn              func(neName string) ([]*db_models.CliConfigBackup, error)
	GetConfigBackupByIdFn            func(id int64) (*db_models.CliConfigBackup, error)
}

func (m *MockStore) Init(cfg config_models.DatabaseConfig) error {
	if m.InitFn != nil {
		return m.InitFn(cfg)
	}
	return nil
}

func (m *MockStore) GetAllUser() ([]*db_models.TblAccount, error) {
	if m.GetAllUserFn != nil {
		return m.GetAllUserFn()
	}
	return nil, nil
}

func (m *MockStore) GetUserByUserName(name string) (*db_models.TblAccount, error) {
	if m.GetUserByUserNameFn != nil {
		return m.GetUserByUserNameFn(name)
	}
	return nil, nil
}

func (m *MockStore) UpdateUser(account *db_models.TblAccount) error {
	if m.UpdateUserFn != nil {
		return m.UpdateUserFn(account)
	}
	return nil
}

func (m *MockStore) AddUser(account *db_models.TblAccount) error {
	if m.AddUserFn != nil {
		return m.AddUserFn(account)
	}
	return nil
}

func (m *MockStore) Ping() error {
	if m.PingFn != nil {
		return m.PingFn()
	}
	return nil
}

func (m *MockStore) UpdateLoginHistory(username, ip string, t time.Time) error {
	if m.UpdateLoginHistoryFn != nil {
		return m.UpdateLoginHistoryFn(username, ip, t)
	}
	return nil
}

func (m *MockStore) SaveHistoryCommand(h db_models.CliOperationHistory) error {
	if m.SaveHistoryCommandFn != nil {
		return m.SaveHistoryCommandFn(h)
	}
	return nil
}

func (m *MockStore) GetCLIUserNeMappingByUserId(userID int64) (*db_models.CliUserNeMapping, error) {
	if m.GetCLIUserNeMappingByUserIdFn != nil {
		return m.GetCLIUserNeMappingByUserIdFn(userID)
	}
	return nil, nil
}

func (m *MockStore) GetNeListById(id int64) ([]*db_models.CliNe, error) {
	if m.GetNeListByIdFn != nil {
		return m.GetNeListByIdFn(id)
	}
	return nil, nil
}

func (m *MockStore) GetCliNeListBySystemType(systemType string) ([]*db_models.CliNe, error) {
	if m.GetCliNeListBySystemTypeFn != nil {
		return m.GetCliNeListBySystemTypeFn(systemType)
	}
	return nil, nil
}

func (m *MockStore) GetCliNeByNeId(id int64) (*db_models.CliNe, error) {
	if m.GetCliNeByNeIdFn != nil {
		return m.GetCliNeByNeIdFn(id)
	}
	return nil, nil
}

func (m *MockStore) CreateCliNe(ne *db_models.CliNe) error {
	if m.CreateCliNeFn != nil {
		return m.CreateCliNeFn(ne)
	}
	return nil
}

func (m *MockStore) UpdateCliNe(ne *db_models.CliNe) error {
	if m.UpdateCliNeFn != nil {
		return m.UpdateCliNeFn(ne)
	}
	return nil
}

func (m *MockStore) DeleteCliNeById(id int64) error {
	if m.DeleteCliNeByIdFn != nil {
		return m.DeleteCliNeByIdFn(id)
	}
	return nil
}

func (m *MockStore) CreateUserNeMapping(mapping *db_models.CliUserNeMapping) error {
	if m.CreateUserNeMappingFn != nil {
		return m.CreateUserNeMappingFn(mapping)
	}
	return nil
}

func (m *MockStore) DeleteUserNeMapping(mapping *db_models.CliUserNeMapping) error {
	if m.DeleteUserNeMappingFn != nil {
		return m.DeleteUserNeMappingFn(mapping)
	}
	return nil
}

func (m *MockStore) DeleteAllUserNeMappingByNeId(neId int64) error {
	if m.DeleteAllUserNeMappingByNeIdFn != nil {
		return m.DeleteAllUserNeMappingByNeIdFn(neId)
	}
	return nil
}

func (m *MockStore) GetNeMonitorById(id int64) (*db_models.CliNeMonitor, error) {
	if m.GetNeMonitorByIdFn != nil {
		return m.GetNeMonitorByIdFn(id)
	}
	return nil, nil
}

func (m *MockStore) DeleteNeMonitorByNeId(neId int64) error {
	if m.DeleteNeMonitorByNeIdFn != nil {
		return m.DeleteNeMonitorByNeIdFn(neId)
	}
	return nil
}

func (m *MockStore) GetAllNeOfUserByUserId(userID int64) ([]*db_models.CliUserNeMapping, error) {
	if m.GetAllNeOfUserByUserIdFn != nil {
		return m.GetAllNeOfUserByUserIdFn(userID)
	}
	return nil, nil
}

func (m *MockStore) DeleteCliNeSlaveByNeId(neId int64) error {
	if m.DeleteCliNeSlaveByNeIdFn != nil {
		return m.DeleteCliNeSlaveByNeIdFn(neId)
	}
	return nil
}

func (m *MockStore) CreateCliNeConfig(cfg *db_models.CliNeConfig) error {
	if m.CreateCliNeConfigFn != nil {
		return m.CreateCliNeConfigFn(cfg)
	}
	return nil
}

func (m *MockStore) GetCliNeConfigByNeId(neId int64) ([]*db_models.CliNeConfig, error) {
	if m.GetCliNeConfigByNeIdFn != nil {
		return m.GetCliNeConfigByNeIdFn(neId)
	}
	return nil, nil
}

func (m *MockStore) GetCliNeConfigById(id int64) (*db_models.CliNeConfig, error) {
	if m.GetCliNeConfigByIdFn != nil {
		return m.GetCliNeConfigByIdFn(id)
	}
	return nil, nil
}

func (m *MockStore) UpdateCliNeConfig(cfg *db_models.CliNeConfig) error {
	if m.UpdateCliNeConfigFn != nil {
		return m.UpdateCliNeConfigFn(cfg)
	}
	return nil
}

func (m *MockStore) DeleteCliNeConfigById(id int64) error {
	if m.DeleteCliNeConfigByIdFn != nil {
		return m.DeleteCliNeConfigByIdFn(id)
	}
	return nil
}

func (m *MockStore) DeleteCliNeConfigByNeId(neId int64) error {
	if m.DeleteCliNeConfigByNeIdFn != nil {
		return m.DeleteCliNeConfigByNeIdFn(neId)
	}
	return nil
}

func (m *MockStore) GetRecentHistory(limit int) ([]db_models.CliOperationHistory, error) {
	if m.GetRecentHistoryFn != nil {
		return m.GetRecentHistoryFn(limit)
	}
	return nil, nil
}

func (m *MockStore) GetRecentHistoryFiltered(limit int, scope, neName string) ([]db_models.CliOperationHistory, error) {
	if m.GetRecentHistoryFilteredFn != nil {
		return m.GetRecentHistoryFilteredFn(limit, scope, neName)
	}
	return nil, nil
}

func (m *MockStore) GetDailyOperationHistory(date time.Time) ([]db_models.CliOperationHistory, error) {
	if m.GetDailyOperationHistoryFn != nil {
		return m.GetDailyOperationHistoryFn(date)
	}
	return nil, nil
}

func (m *MockStore) DeleteHistoryBefore(cutoff time.Time) (int64, error) {
	if m.DeleteHistoryBeforeFn != nil {
		return m.DeleteHistoryBeforeFn(cutoff)
	}
	return 0, nil
}

func (m *MockStore) SaveConfigBackup(b *db_models.CliConfigBackup) error {
	if m.SaveConfigBackupFn != nil {
		return m.SaveConfigBackupFn(b)
	}
	return nil
}

func (m *MockStore) ListConfigBackups(neName string) ([]*db_models.CliConfigBackup, error) {
	if m.ListConfigBackupsFn != nil {
		return m.ListConfigBackupsFn(neName)
	}
	return nil, nil
}

func (m *MockStore) GetConfigBackupById(id int64) (*db_models.CliConfigBackup, error) {
	if m.GetConfigBackupByIdFn != nil {
		return m.GetConfigBackupByIdFn(id)
	}
	return nil, nil
}
