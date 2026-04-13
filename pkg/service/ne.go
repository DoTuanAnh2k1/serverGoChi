package service

import (
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
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

// DeleteNeById cascades: removes all relations before deleting the NE itself.
func DeleteNeById(id int64) error {
	s := store.GetSingleton()
	steps := []struct {
		name string
		fn   func() error
	}{
		{"user-ne mappings", func() error { return s.DeleteAllUserNeMappingByNeId(id) }},
		{"ne monitor", func() error { return s.DeleteNeMonitorByNeId(id) }},
		{"ne config", func() error { return s.DeleteCliNeConfigByNeId(id) }},
		{"ne slave", func() error { return s.DeleteCliNeSlaveByNeId(id) }},
		{"ne", func() error { return s.DeleteCliNeById(id) }},
	}
	for _, step := range steps {
		if err := step.fn(); err != nil {
			logger.Logger.WithField("ne_id", id).Errorf("ne: delete cascade [%s]: %v", step.name, err)
			return err
		}
	}
	return nil
}

func CreateNe(ne *db_models.CliNe) error {
	if err := store.GetSingleton().CreateCliNe(ne); err != nil {
		logger.Logger.WithField("ne_name", ne.Name).Errorf("ne: create: %v", err)
		return err
	}
	return nil
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

func CreateNeConfig(cfg *db_models.CliNeConfig) error {
	if err := store.GetSingleton().CreateCliNeConfig(cfg); err != nil {
		logger.Logger.WithField("ne_id", cfg.NeID).Errorf("ne_config: create: %v", err)
		return err
	}
	return nil
}

func GetNeConfigByNeId(neId int64) ([]*db_models.CliNeConfig, error) {
	list, err := store.GetSingleton().GetCliNeConfigByNeId(neId)
	if err != nil {
		logger.Logger.WithField("ne_id", neId).Errorf("ne_config: get by ne_id: %v", err)
		return nil, err
	}
	return list, nil
}

func GetNeConfigById(id int64) (*db_models.CliNeConfig, error) {
	cfg, err := store.GetSingleton().GetCliNeConfigById(id)
	if err != nil {
		logger.Logger.WithField("id", id).Errorf("ne_config: get by id: %v", err)
		return nil, err
	}
	return cfg, nil
}

func UpdateNeConfig(cfg *db_models.CliNeConfig) error {
	if err := store.GetSingleton().UpdateCliNeConfig(cfg); err != nil {
		logger.Logger.WithField("id", cfg.ID).Errorf("ne_config: update: %v", err)
		return err
	}
	return nil
}

func DeleteNeConfigById(id int64) error {
	if err := store.GetSingleton().DeleteCliNeConfigById(id); err != nil {
		logger.Logger.WithField("id", id).Errorf("ne_config: delete: %v", err)
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
