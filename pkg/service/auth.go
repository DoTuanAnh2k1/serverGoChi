package service

import (
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/bcrypt"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
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

// GetPermissionByUser derives "admin" or "user" from account_type.
// account_type 0 (SuperAdmin) and 1 (Admin) → "admin"; 2 (Normal) → "user".
func GetPermissionByUser(u *db_models.TblAccount) string {
	if u.AccountType <= 1 {
		return "admin"
	}
	return "user"
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
