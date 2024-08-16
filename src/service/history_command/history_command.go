package history_command

import (
	"serverGoChi/models/db_models"
	"serverGoChi/src/log"
	"serverGoChi/src/store"
)

func SaveHistoryCommand(historyCommand db_models.CliOperationHistory) error {
	sto := store.GetSingleton()
	log.Logger.Debug("Save command")
	err := sto.SaveHistoryCommand(historyCommand)
	if err != nil {
		log.Logger.Error("Cant save history command: ", err)
		return err
	}
	return nil
}
