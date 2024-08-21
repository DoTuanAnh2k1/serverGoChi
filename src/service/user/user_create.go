package user

import (
	"serverGoChi/models/db_models"
	"serverGoChi/src/logger"
	"serverGoChi/src/store"
)

func AddUser(user *db_models.TblAccount) error {
	sto := store.GetSingleton()
	err := sto.AddUser(user)
	if err != nil {
		logger.Logger.Error("Failed to create user")
		return err
	}
	return nil
}
