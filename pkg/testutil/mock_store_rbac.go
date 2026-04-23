package testutil

import (
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// Fn fields for the RBAC surface. Kept in a separate file so the primary
// mock_store.go stays manageable. Every method falls back to a sensible
// zero value when the Fn field is unset, so tests that don't exercise RBAC
// can leave these nil.

// ── cli_ne_profile ──

func (m *MockStore) CreateNeProfile(p *db_models.CliNeProfile) error {
	if m.CreateNeProfileFn != nil {
		return m.CreateNeProfileFn(p)
	}
	return nil
}

func (m *MockStore) GetNeProfileById(id int64) (*db_models.CliNeProfile, error) {
	if m.GetNeProfileByIdFn != nil {
		return m.GetNeProfileByIdFn(id)
	}
	return nil, nil
}

func (m *MockStore) GetNeProfileByName(name string) (*db_models.CliNeProfile, error) {
	if m.GetNeProfileByNameFn != nil {
		return m.GetNeProfileByNameFn(name)
	}
	return nil, nil
}

func (m *MockStore) ListNeProfiles() ([]*db_models.CliNeProfile, error) {
	if m.ListNeProfilesFn != nil {
		return m.ListNeProfilesFn()
	}
	return nil, nil
}

func (m *MockStore) UpdateNeProfile(p *db_models.CliNeProfile) error {
	if m.UpdateNeProfileFn != nil {
		return m.UpdateNeProfileFn(p)
	}
	return nil
}

func (m *MockStore) DeleteNeProfileById(id int64) error {
	if m.DeleteNeProfileByIdFn != nil {
		return m.DeleteNeProfileByIdFn(id)
	}
	return nil
}

// ── cli_command_def ──

func (m *MockStore) CreateCommandDef(d *db_models.CliCommandDef) error {
	if m.CreateCommandDefFn != nil {
		return m.CreateCommandDefFn(d)
	}
	return nil
}

func (m *MockStore) GetCommandDefById(id int64) (*db_models.CliCommandDef, error) {
	if m.GetCommandDefByIdFn != nil {
		return m.GetCommandDefByIdFn(id)
	}
	return nil, nil
}

func (m *MockStore) ListCommandDefs(service, neProfile, category string) ([]*db_models.CliCommandDef, error) {
	if m.ListCommandDefsFn != nil {
		return m.ListCommandDefsFn(service, neProfile, category)
	}
	return nil, nil
}

func (m *MockStore) UpdateCommandDef(d *db_models.CliCommandDef) error {
	if m.UpdateCommandDefFn != nil {
		return m.UpdateCommandDefFn(d)
	}
	return nil
}

func (m *MockStore) DeleteCommandDefById(id int64) error {
	if m.DeleteCommandDefByIdFn != nil {
		return m.DeleteCommandDefByIdFn(id)
	}
	return nil
}

// ── cli_command_group ──

func (m *MockStore) CreateCommandGroup(g *db_models.CliCommandGroup) error {
	if m.CreateCommandGroupFn != nil {
		return m.CreateCommandGroupFn(g)
	}
	return nil
}

func (m *MockStore) GetCommandGroupById(id int64) (*db_models.CliCommandGroup, error) {
	if m.GetCommandGroupByIdFn != nil {
		return m.GetCommandGroupByIdFn(id)
	}
	return nil, nil
}

func (m *MockStore) GetCommandGroupByName(name string) (*db_models.CliCommandGroup, error) {
	if m.GetCommandGroupByNameFn != nil {
		return m.GetCommandGroupByNameFn(name)
	}
	return nil, nil
}

func (m *MockStore) ListCommandGroups(service, neProfile string) ([]*db_models.CliCommandGroup, error) {
	if m.ListCommandGroupsFn != nil {
		return m.ListCommandGroupsFn(service, neProfile)
	}
	return nil, nil
}

func (m *MockStore) UpdateCommandGroup(g *db_models.CliCommandGroup) error {
	if m.UpdateCommandGroupFn != nil {
		return m.UpdateCommandGroupFn(g)
	}
	return nil
}

func (m *MockStore) DeleteCommandGroupById(id int64) error {
	if m.DeleteCommandGroupByIdFn != nil {
		return m.DeleteCommandGroupByIdFn(id)
	}
	return nil
}

// ── cli_command_group_mapping ──

func (m *MockStore) AddCommandToGroup(x *db_models.CliCommandGroupMapping) error {
	if m.AddCommandToGroupFn != nil {
		return m.AddCommandToGroupFn(x)
	}
	return nil
}

func (m *MockStore) RemoveCommandFromGroup(x *db_models.CliCommandGroupMapping) error {
	if m.RemoveCommandFromGroupFn != nil {
		return m.RemoveCommandFromGroupFn(x)
	}
	return nil
}

func (m *MockStore) ListCommandsOfGroup(groupId int64) ([]*db_models.CliCommandDef, error) {
	if m.ListCommandsOfGroupFn != nil {
		return m.ListCommandsOfGroupFn(groupId)
	}
	return nil, nil
}

func (m *MockStore) ListGroupsOfCommand(commandId int64) ([]*db_models.CliCommandGroup, error) {
	if m.ListGroupsOfCommandFn != nil {
		return m.ListGroupsOfCommandFn(commandId)
	}
	return nil, nil
}

func (m *MockStore) DeleteAllCommandGroupMappingByGroupId(groupId int64) error {
	if m.DeleteAllCommandGroupMappingByGroupIdFn != nil {
		return m.DeleteAllCommandGroupMappingByGroupIdFn(groupId)
	}
	return nil
}

func (m *MockStore) DeleteAllCommandGroupMappingByCommandId(commandId int64) error {
	if m.DeleteAllCommandGroupMappingByCommandIdFn != nil {
		return m.DeleteAllCommandGroupMappingByCommandIdFn(commandId)
	}
	return nil
}

// ── cli_group_cmd_permission ──

func (m *MockStore) CreateGroupCmdPermission(p *db_models.CliGroupCmdPermission) error {
	if m.CreateGroupCmdPermissionFn != nil {
		return m.CreateGroupCmdPermissionFn(p)
	}
	return nil
}

func (m *MockStore) GetGroupCmdPermissionById(id int64) (*db_models.CliGroupCmdPermission, error) {
	if m.GetGroupCmdPermissionByIdFn != nil {
		return m.GetGroupCmdPermissionByIdFn(id)
	}
	return nil, nil
}

func (m *MockStore) ListGroupCmdPermissions(groupId int64) ([]*db_models.CliGroupCmdPermission, error) {
	if m.ListGroupCmdPermissionsFn != nil {
		return m.ListGroupCmdPermissionsFn(groupId)
	}
	return nil, nil
}

func (m *MockStore) DeleteGroupCmdPermissionById(id int64) error {
	if m.DeleteGroupCmdPermissionByIdFn != nil {
		return m.DeleteGroupCmdPermissionByIdFn(id)
	}
	return nil
}

func (m *MockStore) DeleteAllGroupCmdPermissionByGroupId(groupId int64) error {
	if m.DeleteAllGroupCmdPermissionByGroupIdFn != nil {
		return m.DeleteAllGroupCmdPermissionByGroupIdFn(groupId)
	}
	return nil
}
