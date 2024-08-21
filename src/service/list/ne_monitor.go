package list

import (
	"serverGoChi/models/db_models"
	"serverGoChi/src/logger"
	"serverGoChi/src/store"
)

func GetNeMonitorById(id int64) (*db_models.CliNeMonitor, error) {
	sto := store.GetSingleton()
	cliNeMonitor, err := sto.GetNeMonitorById(id)
	if err != nil {
		logger.Logger.Error("Cannot get cli ne monitor, err: ", err)
		return nil, err
	}
	return cliNeMonitor, nil
}
