package service

import (
	"errors"
	"fmt"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
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

// PurgeUser hard-deletes a user and every related row — the tbl_account row
// itself, its user-group and user-ne mappings, and its password history.
// This is DESTRUCTIVE and irreversible; the legacy `DeleteUser` soft-delete
// (is_enable=false) remains the default in most UIs. SuperAdmin accounts
// are refused — they can only be removed by direct DB access.
//
// Execution order matters: mappings first (their FK constraints reference
// tbl_account), then password history, then the account row last. We best-
// effort each cascade step — log failures but continue — so a stuck mapping
// doesn't leave the tbl_account row orphaned. The final DeleteUserById is
// the authoritative step: its error is returned.
func PurgeUser(accountName string) error {
	sto := store.GetSingleton()
	u, err := sto.GetUserByUserName(accountName)
	if err != nil {
		return err
	}
	if u == nil {
		return fmt.Errorf("user %q not found", accountName)
	}
	if u.AccountType == 0 {
		return errors.New("refusing to purge SuperAdmin account")
	}
	log := logger.Logger.WithField("target", accountName).WithField("user_id", u.AccountID)

	if err := sto.DeleteAllUserGroupMappingByUserId(u.AccountID); err != nil {
		log.Warnf("purge: clear user-group mappings: %v", err)
	}
	if err := sto.DeleteAllUserNeMappingByUserId(u.AccountID); err != nil {
		log.Warnf("purge: clear user-ne mappings: %v", err)
	}
	// Prune history to 0 = delete every row for this user.
	if err := sto.PrunePasswordHistory(u.AccountID, 0); err != nil {
		log.Warnf("purge: clear password history: %v", err)
	}
	if err := sto.DeleteUserById(u.AccountID); err != nil {
		log.Errorf("purge: delete tbl_account row: %v", err)
		return err
	}
	log.Info("purge: user hard-deleted")
	return nil
}
