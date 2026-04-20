package service

import (
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

// ── cli_group CRUD ───────────────────────────────────────────────────────────

func CreateGroup(g *db_models.CliGroup) error {
	if err := store.GetSingleton().CreateGroup(g); err != nil {
		logger.Logger.WithField("group", g.Name).Errorf("group: create: %v", err)
		return err
	}
	return nil
}

func GetGroupById(id int64) (*db_models.CliGroup, error) {
	g, err := store.GetSingleton().GetGroupById(id)
	if err != nil {
		logger.Logger.WithField("group_id", id).Errorf("group: get by id: %v", err)
		return nil, err
	}
	return g, nil
}

func GetGroupByName(name string) (*db_models.CliGroup, error) {
	g, err := store.GetSingleton().GetGroupByName(name)
	if err != nil {
		logger.Logger.WithField("group", name).Errorf("group: get by name: %v", err)
		return nil, err
	}
	return g, nil
}

func GetAllGroups() ([]*db_models.CliGroup, error) {
	list, err := store.GetSingleton().GetAllGroups()
	if err != nil {
		logger.Logger.Errorf("group: list all: %v", err)
		return nil, err
	}
	return list, nil
}

func UpdateGroup(g *db_models.CliGroup) error {
	if err := store.GetSingleton().UpdateGroup(g); err != nil {
		logger.Logger.WithField("group_id", g.ID).Errorf("group: update: %v", err)
		return err
	}
	return nil
}

// DeleteGroupById cascades: removes user-group and group-ne mappings before the group itself.
func DeleteGroupById(id int64) error {
	s := store.GetSingleton()
	if err := s.DeleteAllUserGroupMappingByGroupId(id); err != nil {
		logger.Logger.WithField("group_id", id).Errorf("group: cascade user-group: %v", err)
		return err
	}
	if err := s.DeleteAllGroupNeMappingByGroupId(id); err != nil {
		logger.Logger.WithField("group_id", id).Errorf("group: cascade group-ne: %v", err)
		return err
	}
	if err := s.DeleteGroupById(id); err != nil {
		logger.Logger.WithField("group_id", id).Errorf("group: delete: %v", err)
		return err
	}
	return nil
}

// ── user ↔ group ─────────────────────────────────────────────────────────────

func AssignUserToGroup(userId, groupId int64) error {
	m := &db_models.CliUserGroupMapping{UserID: userId, GroupID: groupId}
	if err := store.GetSingleton().CreateUserGroupMapping(m); err != nil {
		logger.Logger.WithField("user_id", userId).WithField("group_id", groupId).Errorf("group: assign user: %v", err)
		return err
	}
	return nil
}

func UnassignUserFromGroup(userId, groupId int64) error {
	m := &db_models.CliUserGroupMapping{UserID: userId, GroupID: groupId}
	if err := store.GetSingleton().DeleteUserGroupMapping(m); err != nil {
		logger.Logger.WithField("user_id", userId).WithField("group_id", groupId).Errorf("group: unassign user: %v", err)
		return err
	}
	return nil
}

func GetGroupsOfUser(userId int64) ([]*db_models.CliUserGroupMapping, error) {
	list, err := store.GetSingleton().GetAllGroupsOfUser(userId)
	if err != nil {
		logger.Logger.WithField("user_id", userId).Errorf("group: list of user: %v", err)
		return nil, err
	}
	return list, nil
}

func GetUsersOfGroup(groupId int64) ([]*db_models.CliUserGroupMapping, error) {
	list, err := store.GetSingleton().GetAllUsersOfGroup(groupId)
	if err != nil {
		logger.Logger.WithField("group_id", groupId).Errorf("group: users of group: %v", err)
		return nil, err
	}
	return list, nil
}

// ── group ↔ NE ───────────────────────────────────────────────────────────────

func AssignNeToGroup(groupId, neId int64) error {
	m := &db_models.CliGroupNeMapping{GroupID: groupId, TblNeID: neId}
	if err := store.GetSingleton().CreateGroupNeMapping(m); err != nil {
		logger.Logger.WithField("group_id", groupId).WithField("ne_id", neId).Errorf("group: assign ne: %v", err)
		return err
	}
	return nil
}

func UnassignNeFromGroup(groupId, neId int64) error {
	m := &db_models.CliGroupNeMapping{GroupID: groupId, TblNeID: neId}
	if err := store.GetSingleton().DeleteGroupNeMapping(m); err != nil {
		logger.Logger.WithField("group_id", groupId).WithField("ne_id", neId).Errorf("group: unassign ne: %v", err)
		return err
	}
	return nil
}

func GetNesOfGroup(groupId int64) ([]*db_models.CliGroupNeMapping, error) {
	list, err := store.GetSingleton().GetAllNesOfGroup(groupId)
	if err != nil {
		logger.Logger.WithField("group_id", groupId).Errorf("group: nes of group: %v", err)
		return nil, err
	}
	return list, nil
}

func GetGroupsOfNe(neId int64) ([]*db_models.CliGroupNeMapping, error) {
	list, err := store.GetSingleton().GetAllGroupsOfNe(neId)
	if err != nil {
		logger.Logger.WithField("ne_id", neId).Errorf("group: groups of ne: %v", err)
		return nil, err
	}
	return list, nil
}

// GetAllNeIdsOfUser returns the union of NE ids reachable by the user —
// directly via cli_user_ne_mapping plus indirectly via any group they belong to.
// Deduplicated.
func GetAllNeIdsOfUser(userId int64) ([]int64, error) {
	s := store.GetSingleton()
	seen := map[int64]struct{}{}

	directs, err := s.GetAllNeOfUserByUserId(userId)
	if err != nil {
		return nil, err
	}
	for _, m := range directs {
		seen[m.TblNeID] = struct{}{}
	}

	groupMaps, err := s.GetAllGroupsOfUser(userId)
	if err != nil {
		return nil, err
	}
	for _, gm := range groupMaps {
		nes, err := s.GetAllNesOfGroup(gm.GroupID)
		if err != nil {
			return nil, err
		}
		for _, m := range nes {
			seen[m.TblNeID] = struct{}{}
		}
	}

	out := make([]int64, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out, nil
}
