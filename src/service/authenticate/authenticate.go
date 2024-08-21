package authenticate

import (
	"serverGoChi/models/db_models"
	"serverGoChi/src/logger"
	"serverGoChi/src/store"
	"serverGoChi/src/utils/bcrypt"
	"time"
)

func UpdateLoginHistory(username, ipAddress string) error {
	timeLogin := time.Now()
	sto := store.GetSingleton()
	err := sto.UpdateLoginHistory(username, ipAddress, timeLogin)
	if err != nil {
		logger.Logger.Error("Cant insert to database: ", err)
		return err
	}
	return nil
}

func GetTblIdByUserId(userId int64) (int64, error) {
	sto := store.GetSingleton()
	cliUserNeMapping, err := sto.GetCLIUserNeMappingByUserId(userId)
	if err != nil {
		logger.Logger.Error("Cant get from database: ", err)
		return 0, err
	}
	return cliUserNeMapping.TblNeID, nil
}

func GetNeListById(id int64) ([]*db_models.CliNe, error) {
	sto := store.GetSingleton()
	cliNeList, err := sto.GetNeListById(id)
	if err != nil {
		logger.Logger.Error("Cant get ne list from database: ", err)
		return nil, err
	}
	return cliNeList, nil
}

func GetRolesById(userId int64) (string, error) {
	sto := store.GetSingleton()
	roleList, err := sto.GetRolesById(userId)
	if err != nil {
		logger.Logger.Error("Cant get role list from database: ", err)
		return "", err
	}
	roleToString := ""
	for _, v := range roleList {
		roleToString = roleToString + " " + v.Permission
	}
	return roleToString, nil
}

func Authenticate(username, password string) (bool, error, int64) {
	sto := store.GetSingleton()
	user, err := sto.GetUserByUserName(username)
	if err != nil {
		logger.Logger.Error("Cant user by username from database: ", err)
		return false, err, -1
	}
	if bcrypt.Matches(username+password, user.Password) {
		return true, nil, user.AccountID
	}
	return false, nil, -1
}
