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
	"serverGoChi/src/service/user"
	"time"
)

func HandlerUserDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var userDeleteReq UserDeleteReq
	err := json.NewDecoder(r.Body).Decode(&userDeleteReq)
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
		CmdName:     fmt.Sprintf("authorize-user set username %v permission %v", userDeleteReq.Username, userDeleteReq.Permission),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}

	log.Logger.Info("Handler User Delete with username: ", userMiddleware.Username)
	u, err := user.GetUserByUserName(userMiddleware.Username)
	if err != nil {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err1 := history_command.SaveHistoryCommand(logOperationHistory)
		if err1 != nil {
			log.Logger.Error("Cannot save command to db: ", err1)
		}

		log.Logger.Error("Error to get user from db: ", err)
		response.InternalError(w, "Error to get user from db")
		return
	}

	if u == nil {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cannot save command to db: ", err)
		}

		log.Logger.Info("Not found user")
		response.NotFound(w, "User not found")
		return
	}
	log.Logger.Info("Found User: ", u.AccountName)

	roles, err := authorize.GetAllUserRolesMappingById(u.AccountID)
	if err != nil {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cannot save command to db: ", err)
		}

		log.Logger.Error("Error to get role of user from db: ", err)
		response.InternalError(w, "Error to get role of user from db")
		return
	}
	log.Logger.Info("User roles: ", roles)

	for _, role := range roles {
		if role.Permission != userDeleteReq.Permission {
			continue
		}
		err = authorize.DeleteUserRole(&db_models.CliRoleUserMapping{
			UserID:     u.AccountID,
			Permission: userDeleteReq.Permission,
		})
		if err != nil {
			logOperationHistory.ExecutedTime = time.Now()
			logOperationHistory.Result = "failure"
			err1 := history_command.SaveHistoryCommand(logOperationHistory)
			if err1 != nil {
				log.Logger.Error("Cannot save command to db: ", err1)
			}

			log.Logger.Error("Cannot Delete user role: ", err)
			response.InternalError(w, "Cannot Delete User Role")
			return
		}
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "success"
		err := history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cannot save command to db: ", err)
		}
		log.Logger.Info("Delete User Role Success")
		response.Write(w, http.StatusOK, "Delete Success")
		return
	}

	logOperationHistory.ExecutedTime = time.Now()
	logOperationHistory.Result = "failure"
	err = history_command.SaveHistoryCommand(logOperationHistory)
	if err != nil {
		log.Logger.Error("Cannot save command to db: ", err)
	}
	log.Logger.Info("Cannot find permission of user")
	response.NotFound(w, "Cannot find permission of user")
	return
}
