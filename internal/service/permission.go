package service

import (
	"go-aa-server/internal/logger"
	"go-aa-server/internal/store"
	"go-aa-server/models/db_models"
)

func IsExistCliRole(cliRole *db_models.CliRole) (bool, error) {
	found, err := store.GetSingleton().GetCliRole(cliRole)
	if err != nil {
		logger.Logger.Errorf("permission: check role exists: %v", err)
		return false, err
	}
	return found != nil, nil
}

func CreateCliRole(cliRole *db_models.CliRole) error {
	if err := store.GetSingleton().CreateCliRole(cliRole); err != nil {
		logger.Logger.Errorf("permission: create role: %v", err)
		return err
	}
	return nil
}

func DeleteCliRole(cliRole *db_models.CliRole) error {
	if err := store.GetSingleton().DeleteCliRole(cliRole); err != nil {
		logger.Logger.Errorf("permission: delete role: %v", err)
		return err
	}
	return nil
}

func GetAllCliRoles() ([]*db_models.CliRole, error) {
	list, err := store.GetSingleton().GetAllCliRole()
	if err != nil {
		logger.Logger.Errorf("permission: get all roles: %v", err)
		return nil, err
	}
	return list, nil
}

func GetAllUserRolesMappingById(id int64) ([]*db_models.CliRoleUserMapping, error) {
	roles, err := store.GetSingleton().GetRolesById(id)
	if err != nil {
		logger.Logger.WithField("user_id", id).Errorf("permission: get user roles: %v", err)
		return nil, err
	}
	return roles, nil
}

func AddUserRole(role *db_models.CliRoleUserMapping) error {
	if err := store.GetSingleton().AddRole(role); err != nil {
		logger.Logger.WithField("user_id", role.UserID).Errorf("permission: add role %q: %v", role.Permission, err)
		return err
	}
	return nil
}

func DeleteUserRole(role *db_models.CliRoleUserMapping) error {
	if err := store.GetSingleton().DeleteRole(role); err != nil {
		logger.Logger.WithField("user_id", role.UserID).Errorf("permission: delete role %q: %v", role.Permission, err)
		return err
	}
	return nil
}
