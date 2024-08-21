package list

import (
	"serverGoChi/models/db_models"
	"serverGoChi/src/log"
	"serverGoChi/src/store"
)

func GetNeMonitorById(id int64) (*db_models.CliNeMonitor, error) {
	sto := store.GetSingleton()
	cliNeMonitor, err := sto.GetNeMonitorById(id)
	if err != nil {
		log.Logger.Error("Cannot get cli ne monitor, err: ", err)
		return nil, err
	}
	return cliNeMonitor, nil
}
