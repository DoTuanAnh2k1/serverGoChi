package authorize

import (
	"encoding/json"
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

func HandlerPermissionDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var cliRole db_models.CliRole
	err := json.NewDecoder(r.Body).Decode(&cliRole)
	if err != nil {
		log.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}

	userMiddleware, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		log.Logger.Error("Error to get user from token key")
		response.InternalError(w, "Internal Server Error")
		return
	}

	logOperationHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authorize-permission set permission %v scope %v ne %v include type %v path %v", cliRole.Permission, cliRole.Scope, cliRole.NeType, cliRole.IncludeType, cliRole.Path),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}

	isExist, err := authorize.IsExistCliRole(cliRole)
	if err != nil {
		log.Logger.Error("Cant check is exist cli role: ", err)
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err1 := history_command.SaveHistoryCommand(logOperationHistory)
		if err1 != nil {
			log.Logger.Error("Cant save command to db: ", err1)
		}
		response.Write(w, http.StatusInternalServerError, "Cant check is exist cli role")
		return
	}
	if !isExist {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err := history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}
		response.Write(w, http.StatusNotModified, "")
	} else {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "success"
		err := history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}
		err = authorize.DeleteCliRole(cliRole)
		if err != nil {
			log.Logger.Error("Error Delete cli role: ", err)
			response.Write(w, http.StatusInternalServerError, "Error Delete cli role")
			return
		}
		response.Success(w, "")
	}
}
