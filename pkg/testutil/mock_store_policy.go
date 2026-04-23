package testutil

import (
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// Mock impl + Fn fields for password policy, password history, and mgt
// permission. Kept separate from mock_store.go to leave the primary mock
// file manageable.

func (m *MockStore) CreatePasswordPolicy(p *db_models.CliPasswordPolicy) error {
	if m.CreatePasswordPolicyFn != nil {
		return m.CreatePasswordPolicyFn(p)
	}
	return nil
}
func (m *MockStore) GetPasswordPolicyById(id int64) (*db_models.CliPasswordPolicy, error) {
	if m.GetPasswordPolicyByIdFn != nil {
		return m.GetPasswordPolicyByIdFn(id)
	}
	return nil, nil
}
func (m *MockStore) GetPasswordPolicyByName(name string) (*db_models.CliPasswordPolicy, error) {
	if m.GetPasswordPolicyByNameFn != nil {
		return m.GetPasswordPolicyByNameFn(name)
	}
	return nil, nil
}
func (m *MockStore) ListPasswordPolicies() ([]*db_models.CliPasswordPolicy, error) {
	if m.ListPasswordPoliciesFn != nil {
		return m.ListPasswordPoliciesFn()
	}
	return nil, nil
}
func (m *MockStore) UpdatePasswordPolicy(p *db_models.CliPasswordPolicy) error {
	if m.UpdatePasswordPolicyFn != nil {
		return m.UpdatePasswordPolicyFn(p)
	}
	return nil
}
func (m *MockStore) DeletePasswordPolicyById(id int64) error {
	if m.DeletePasswordPolicyByIdFn != nil {
		return m.DeletePasswordPolicyByIdFn(id)
	}
	return nil
}

func (m *MockStore) AppendPasswordHistory(h *db_models.CliPasswordHistory) error {
	if m.AppendPasswordHistoryFn != nil {
		return m.AppendPasswordHistoryFn(h)
	}
	return nil
}
func (m *MockStore) GetRecentPasswordHistory(userID int64, limit int) ([]*db_models.CliPasswordHistory, error) {
	if m.GetRecentPasswordHistoryFn != nil {
		return m.GetRecentPasswordHistoryFn(userID, limit)
	}
	return nil, nil
}
func (m *MockStore) PrunePasswordHistory(userID int64, keep int) error {
	if m.PrunePasswordHistoryFn != nil {
		return m.PrunePasswordHistoryFn(userID, keep)
	}
	return nil
}

func (m *MockStore) CreateMgtPermission(p *db_models.CliGroupMgtPermission) error {
	if m.CreateMgtPermissionFn != nil {
		return m.CreateMgtPermissionFn(p)
	}
	return nil
}
func (m *MockStore) ListMgtPermissions(groupID int64) ([]*db_models.CliGroupMgtPermission, error) {
	if m.ListMgtPermissionsFn != nil {
		return m.ListMgtPermissionsFn(groupID)
	}
	return nil, nil
}
func (m *MockStore) DeleteMgtPermissionById(id int64) error {
	if m.DeleteMgtPermissionByIdFn != nil {
		return m.DeleteMgtPermissionByIdFn(id)
	}
	return nil
}
func (m *MockStore) DeleteAllMgtPermissionByGroupId(groupID int64) error {
	if m.DeleteAllMgtPermissionByGroupIdFn != nil {
		return m.DeleteAllMgtPermissionByGroupIdFn(groupID)
	}
	return nil
}
