package service

import (
	"go-aa-server/internal/logger"
	"go-aa-server/internal/store"
	"go-aa-server/models/db_models"
)

func GetNeListBySystemType(systemType string) ([]*db_models.CliNe, error) {
	list, err := store.GetSingleton().GetCliNeListBySystemType(systemType)
	if err != nil {
		logger.Logger.WithField("system_type", systemType).Errorf("ne: get list by system type: %v", err)
		return nil, err
	}
	return list, nil
}

func GetNeByNeId(id int64) (*db_models.CliNe, error) {
	ne, err := store.GetSingleton().GetCliNeByNeId(id)
	if err != nil {
		logger.Logger.WithField("ne_id", id).Errorf("ne: get by id: %v", err)
		return nil, err
	}
	return ne, nil
}

func GetAllCliNeOfUserByUserId(id int64) ([]*db_models.CliUserNeMapping, error) {
	list, err := store.GetSingleton().GetAllNeOfUserByUserId(id)
	if err != nil {
		logger.Logger.WithField("user_id", id).Errorf("ne: get user-ne mappings: %v", err)
		return nil, err
	}
	return list, nil
}

func AddUserCliNe(m *db_models.CliUserNeMapping) error {
	if err := store.GetSingleton().CreateUserNeMapping(m); err != nil {
		logger.Logger.WithField("user_id", m.UserID).WithField("ne_id", m.TblNeID).Errorf("ne: add user-ne mapping: %v", err)
		return err
	}
	return nil
}

func DeleteCliNe(m *db_models.CliUserNeMapping) error {
	if err := store.GetSingleton().DeleteUserNeMapping(m); err != nil {
		logger.Logger.WithField("user_id", m.UserID).WithField("ne_id", m.TblNeID).Errorf("ne: delete user-ne mapping: %v", err)
		return err
	}
	return nil
}

func GetNeMonitorById(id int64) (*db_models.CliNeMonitor, error) {
	monitor, err := store.GetSingleton().GetNeMonitorById(id)
	if err != nil {
		logger.Logger.WithField("ne_id", id).Errorf("ne: get monitor: %v", err)
		return nil, err
	}
	return monitor, nil
}
