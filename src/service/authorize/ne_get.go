package authorize

import (
	"serverGoChi/models/db_models"
	"serverGoChi/src/log"
	"serverGoChi/src/store"
)

func GetNeListBySystemType(systemType string) ([]*db_models.CliNe, error) {
	sto := store.GetSingleton()
	cliNeList, err := sto.GetCliNeListBySystemType(systemType)
	if err != nil {
		log.Logger.Error("Cant get cli ne list, err: ", err)
		return nil, err
	}
	return cliNeList, nil
}

func GetNeByNeId(id int64) (*db_models.CliNe, error) {
	sto := store.GetSingleton()
	cliNe, err := sto.GetCliNeByNeId(id)
	if err != nil {
		log.Logger.Error("Cant get cli ne, err: ", err)
		return nil, err
	}
	return cliNe, nil
}
