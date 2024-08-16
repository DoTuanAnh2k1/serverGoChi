package user

import (
	"serverGoChi/models/db_models"
	"serverGoChi/src/log"
	"serverGoChi/src/store"
)

func UpdateUser(user db_models.TblAccount) error {
	sto := store.GetSingleton()
	err := sto.UpdateUser(user)
	if err != nil {
		log.Logger.Error("Failed to Update user")
		return err
	}
	return nil
}
