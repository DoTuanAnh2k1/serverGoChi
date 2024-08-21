package user

import (
	"serverGoChi/models/db_models"
	"serverGoChi/src/log"
	"serverGoChi/src/store"
)

func GetAllUser() ([]*db_models.TblAccount, error) {
	sto := store.GetSingleton()
	userList, err := sto.GetAllUser()
	if err != nil {
		log.Logger.Error("Failed to get all user")
		return nil, err
	}
	return userList, nil
}
