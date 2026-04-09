package service

import (
	"strings"
	"time"

	"go-aa-server/internal/bcrypt"
	"go-aa-server/internal/logger"
	"go-aa-server/internal/store"
	"go-aa-server/models/db_models"
)

func UpdateLoginHistory(username, ipAddress string) error {
	err := store.GetSingleton().UpdateLoginHistory(username, ipAddress, time.Now())
	if err != nil {
		logger.Logger.WithField("user", username).Errorf("auth: update login history: %v", err)
		return err
	}
	return nil
}

func GetTblIdByUserId(userId int64) (int64, error) {
	mapping, err := store.GetSingleton().GetCLIUserNeMappingByUserId(userId)
	if err != nil {
		logger.Logger.WithField("user_id", userId).Errorf("auth: get user-ne mapping: %v", err)
		return 0, err
	}
	return mapping.TblNeID, nil
}

func GetNeListById(id int64) ([]*db_models.CliNe, error) {
	list, err := store.GetSingleton().GetNeListById(id)
	if err != nil {
		logger.Logger.WithField("tbl_ne_id", id).Errorf("auth: get ne list: %v", err)
		return nil, err
	}
	return list, nil
}

func GetRolesById(userId int64) (string, error) {
	roleList, err := store.GetSingleton().GetRolesById(userId)
	if err != nil {
		logger.Logger.WithField("user_id", userId).Errorf("auth: get roles: %v", err)
		return "", err
	}
	var perms []string
	for _, r := range roleList {
		perms = append(perms, r.Permission)
	}
	return strings.Join(perms, " "), nil
}

func Authenticate(username, password string) (bool, error, int64) {
	u, err := store.GetSingleton().GetUserByUserName(username)
	if err != nil {
		logger.Logger.WithField("user", username).Errorf("auth: get user: %v", err)
		return false, err, -1
	}
	if !u.IsEnable {
		logger.Logger.WithField("user", username).Warn("auth: login attempt on disabled account")
		return false, nil, -1
	}
	if !bcrypt.Matches(username+password, u.Password) {
		logger.Logger.WithField("user", username).Warn("auth: wrong password")
		return false, nil, -1
	}
	return true, nil, u.AccountID
}
