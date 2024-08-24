package authorize

import (
	"serverGoChi/models/db_models"
	"serverGoChi/src/logger"
	"serverGoChi/src/store"
)

func GetUserByName(name string) (*db_models.TblAccount, error) {
	sto := store.GetSingleton()
	user, err := sto.GetUserByUserName(name)
	if err != nil {
		logger.Logger.Error("Cannot get user: ", err)
		return user, err
	}
	return user, nil
}

func IsExistCliRole(cliRole *db_models.CliRole) (bool, error) {
	sto := store.GetSingleton()
	cliRoleNew, err := sto.GetCliRole(cliRole)
	if err != nil {
		logger.Logger.Error("Cannot get user: ", err)
		return false, err
	}
	if cliRoleNew != nil {
		return true, nil
	}
	return false, nil
}

func CreateCliRole(cliRole *db_models.CliRole) error {
	sto := store.GetSingleton()
	err := sto.CreateCliRole(cliRole)
	if err != nil {
		logger.Logger.Error("Cannot create cli role: ", err)
		return err
	}
	return nil
}

func DeleteCliRole(cliRole *db_models.CliRole) error {
	sto := store.GetSingleton()
	err := sto.DeleteCliRole(cliRole)
	if err != nil {
		logger.Logger.Error("Cannot delete cli role: ", err)
		return err
	}
	return nil
}

func GetAllCliRoles() ([]*db_models.CliRole, error) {
	sto := store.GetSingleton()
	cliRoleList, err := sto.GetAllCliRole()
	if err != nil {
		logger.Logger.Error("Cannot get cli role list: ", err)
		return nil, err
	}
	return cliRoleList, nil
}

func GetAllUserRolesMappingById(id int64) ([]*db_models.CliRoleUserMapping, error) {
	sto := store.GetSingleton()
	roles, err := sto.GetRolesById(id)
	if err != nil {
		logger.Logger.Error("Cannot get role list: ", err)
		return nil, err
	}
	return roles, nil
}

func AddUserRole(role *db_models.CliRoleUserMapping) error {
	sto := store.GetSingleton()
	err := sto.AddRole(role)
	if err != nil {
		logger.Logger.Error("Cannot add role: ", err)
		return err
	}
	return nil
}

func DeleteUserRole(role *db_models.CliRoleUserMapping) error {
	sto := store.GetSingleton()
	err := sto.DeleteRole(role)
	if err != nil {
		logger.Logger.Error("Cannot delete role: ", err)
		return err
	}
	return nil
}

func AddUserCliNe(cliUserNe *db_models.CliUserNeMapping) error {
	sto := store.GetSingleton()
	err := sto.CreateUserNeMapping(cliUserNe)
	if err != nil {
		logger.Logger.Error("Cannot Add Cli Ne to User: ", err)
		return err
	}
	return nil
}

func DeleteCliNe(cliUserNe *db_models.CliUserNeMapping) error {
	sto := store.GetSingleton()
	err := sto.DeleteUserNeMapping(cliUserNe)
	if err != nil {
		logger.Logger.Error("Cannot Delete Cli Ne to User: ", err)
		return err
	}
	return nil
}
