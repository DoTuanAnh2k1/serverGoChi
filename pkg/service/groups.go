package service

import (
	"errors"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
)

var (
	ErrGroupNotFound = errors.New("group: not found")
	ErrGroupExists   = errors.New("group: name already taken")
)

// ── NE Access Group ──

func CreateNeAccessGroup(g *db_models.NeAccessGroup) error {
	existing, err := store.GetSingleton().GetNeAccessGroupByName(g.Name)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrGroupExists
	}
	return store.GetSingleton().CreateNeAccessGroup(g)
}

func GetNeAccessGroup(id int64) (*db_models.NeAccessGroup, error) {
	g, err := store.GetSingleton().GetNeAccessGroupByID(id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrGroupNotFound
	}
	return g, nil
}

func ListNeAccessGroups() ([]*db_models.NeAccessGroup, error) {
	return store.GetSingleton().ListNeAccessGroups()
}

func UpdateNeAccessGroup(g *db_models.NeAccessGroup) error {
	return store.GetSingleton().UpdateNeAccessGroup(g)
}

func DeleteNeAccessGroup(id int64) error {
	return store.GetSingleton().DeleteNeAccessGroupByID(id)
}

func AssignUserToNeAccessGroup(groupID, userID int64) error {
	return store.GetSingleton().AddUserToNeAccessGroup(groupID, userID)
}

func UnassignUserFromNeAccessGroup(groupID, userID int64) error {
	return store.GetSingleton().RemoveUserFromNeAccessGroup(groupID, userID)
}

func ListUsersInNeAccessGroup(groupID int64) ([]int64, error) {
	return store.GetSingleton().ListUsersInNeAccessGroup(groupID)
}

func ListNeAccessGroupsOfUser(userID int64) ([]int64, error) {
	return store.GetSingleton().ListNeAccessGroupsOfUser(userID)
}

func AssignNeToNeAccessGroup(groupID, neID int64) error {
	return store.GetSingleton().AddNeToNeAccessGroup(groupID, neID)
}

func UnassignNeFromNeAccessGroup(groupID, neID int64) error {
	return store.GetSingleton().RemoveNeFromNeAccessGroup(groupID, neID)
}

func ListNEsInNeAccessGroup(groupID int64) ([]int64, error) {
	return store.GetSingleton().ListNEsInNeAccessGroup(groupID)
}

func ListNeAccessGroupsOfNE(neID int64) ([]int64, error) {
	return store.GetSingleton().ListNeAccessGroupsOfNE(neID)
}

// ── Cmd Exec Group ──

func CreateCmdExecGroup(g *db_models.CmdExecGroup) error {
	existing, err := store.GetSingleton().GetCmdExecGroupByName(g.Name)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrGroupExists
	}
	return store.GetSingleton().CreateCmdExecGroup(g)
}

func GetCmdExecGroup(id int64) (*db_models.CmdExecGroup, error) {
	g, err := store.GetSingleton().GetCmdExecGroupByID(id)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrGroupNotFound
	}
	return g, nil
}

func ListCmdExecGroups() ([]*db_models.CmdExecGroup, error) {
	return store.GetSingleton().ListCmdExecGroups()
}

func UpdateCmdExecGroup(g *db_models.CmdExecGroup) error {
	return store.GetSingleton().UpdateCmdExecGroup(g)
}

func DeleteCmdExecGroup(id int64) error {
	return store.GetSingleton().DeleteCmdExecGroupByID(id)
}

func AssignUserToCmdExecGroup(groupID, userID int64) error {
	return store.GetSingleton().AddUserToCmdExecGroup(groupID, userID)
}

func UnassignUserFromCmdExecGroup(groupID, userID int64) error {
	return store.GetSingleton().RemoveUserFromCmdExecGroup(groupID, userID)
}

func ListUsersInCmdExecGroup(groupID int64) ([]int64, error) {
	return store.GetSingleton().ListUsersInCmdExecGroup(groupID)
}

func ListCmdExecGroupsOfUser(userID int64) ([]int64, error) {
	return store.GetSingleton().ListCmdExecGroupsOfUser(userID)
}

func AssignCommandToCmdExecGroup(groupID, commandID int64) error {
	return store.GetSingleton().AddCommandToCmdExecGroup(groupID, commandID)
}

func UnassignCommandFromCmdExecGroup(groupID, commandID int64) error {
	return store.GetSingleton().RemoveCommandFromCmdExecGroup(groupID, commandID)
}

func ListCommandsInCmdExecGroup(groupID int64) ([]int64, error) {
	return store.GetSingleton().ListCommandsInCmdExecGroup(groupID)
}

func ListCmdExecGroupsOfCommand(commandID int64) ([]int64, error) {
	return store.GetSingleton().ListCmdExecGroupsOfCommand(commandID)
}
