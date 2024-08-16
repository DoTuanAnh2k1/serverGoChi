package user

import (
	"serverGoChi/models/db_models"
	"serverGoChi/src/log"
	"serverGoChi/src/store"
)

func GetUserByUserName(name string) (*db_models.TblAccount, error) {
	sto := store.GetSingleton()
	user, err := sto.GetUserByUserName(name)
	if err != nil {
		log.Logger.Error("Failed to get user")
		return nil, err
	}
	// return the user if found, otherwise nil if not found.
	// This function returns a pointer to a db_models.TblAccount struct.
	// This is the struct defined in the db_models package.
	// It contains fields for the user's ID, name, and email.
	// It also has a field for the user's password, which is hashed for security.
	// The password field is not directly accessible from this function.
	// Instead, you must use the Get
	return user, nil
}
