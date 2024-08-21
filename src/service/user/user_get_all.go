package user

import (
	"serverGoChi/models/db_models"
	"serverGoChi/src/logger"
	"serverGoChi/src/store"
)

func GetAllUser() ([]*db_models.TblAccount, error) {
	sto := store.GetSingleton()
	userList, err := sto.GetAllUser()
	if err != nil {
		logger.Logger.Error("Failed to get all user")
		return nil, err
	}
	return userList, nil
}
