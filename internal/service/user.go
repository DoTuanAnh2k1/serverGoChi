package service

import (
	"go-aa-server/internal/logger"
	"go-aa-server/internal/store"
	"go-aa-server/models/db_models"
)

func AddUser(u *db_models.TblAccount) error {
	if err := store.GetSingleton().AddUser(u); err != nil {
		logger.Logger.WithField("user", u.AccountName).Errorf("user: create: %v", err)
		return err
	}
	return nil
}

func UpdateUser(u *db_models.TblAccount) error {
	if err := store.GetSingleton().UpdateUser(u); err != nil {
		logger.Logger.WithField("user", u.AccountName).Errorf("user: update: %v", err)
		return err
	}
	return nil
}

func GetUserByUserName(name string) (*db_models.TblAccount, error) {
	u, err := store.GetSingleton().GetUserByUserName(name)
	if err != nil {
		logger.Logger.WithField("user", name).Errorf("user: get by username: %v", err)
		return nil, err
	}
	return u, nil
}

func GetAllUser() ([]*db_models.TblAccount, error) {
	list, err := store.GetSingleton().GetAllUser()
	if err != nil {
		logger.Logger.Errorf("user: get all: %v", err)
		return nil, err
	}
	return list, nil
}
