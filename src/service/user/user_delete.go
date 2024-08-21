package user

import (
	"serverGoChi/models/db_models"
	"serverGoChi/src/logger"
	"serverGoChi/src/store"
)

func UpdateUser(user *db_models.TblAccount) error {
	sto := store.GetSingleton()
	err := sto.UpdateUser(user)
	if err != nil {
		logger.Logger.Error("Failed to Update user")
		return err
	}
	return nil
}
