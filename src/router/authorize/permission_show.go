package authorize

import (
	"fmt"
	"net/http"
	"serverGoChi/models/db_models"
	"serverGoChi/src/log"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authorize"
	"serverGoChi/src/service/history_command"
	"time"
)

func HandlerPermissionShow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userMiddleware, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		log.Logger.Error("Error to get user from token key")
		response.InternalError(w, "Internal Server Error")
		return
	}

	logOperationHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authorize-permission show"),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}

	cliRoleList, err := authorize.GetAllCliRoles()
	if err != nil {
		log.Logger.Error("Cant get cli role list: ", err)
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err := history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}
		response.Write(w, http.StatusInternalServerError, "Cant get cli role list")
		return
	}

	if cliRoleList == nil || len(cliRoleList) == 0 {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err := history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}
		response.Write(w, http.StatusNotFound, "cli role list empty")
		return
	}

	logOperationHistory.ExecutedTime = time.Now()
	logOperationHistory.Result = "success"
	err = history_command.SaveHistoryCommand(logOperationHistory)
	if err != nil {
		log.Logger.Error("Cant save command to db: ", err)
	}
	response.Write(w, http.StatusFound, cliRoleList)
	return
}
