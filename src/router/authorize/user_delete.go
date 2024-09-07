package authorize

import (
	"encoding/json"
	"fmt"
	"net/http"
	"serverGoChi/models/db_models"
	"serverGoChi/src/logger"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authorize"
	"serverGoChi/src/service/history_command"
	"serverGoChi/src/service/user"
	"time"
)

func HandlerUserDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var userDeleteReq UserDeleteReq
	err := json.NewDecoder(r.Body).Decode(&userDeleteReq)
	if err != nil {
		logger.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}

	userMiddleware, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		logger.Logger.Error("Error to get user from token key")
		response.InternalError(w, "Internal Server Error")
		return
	}

	loggerOperationHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authorize-user set username %v permission %v", userDeleteReq.Username, userDeleteReq.Permission),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}

	logger.Logger.Info("Handler User Delete with username: ", userDeleteReq.Username)
	u, err := user.GetUserByUserName(userDeleteReq.Username)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err1 := history_command.SaveHistoryCommand(loggerOperationHistory)
		if err1 != nil {
			logger.Logger.Error("Cannot save command to db: ", err1)
		}

		logger.Logger.Error("Error to get user from db: ", err)
		response.InternalError(w, "Error to get user from db")
		return
	}

	if u == nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cannot save command to db: ", err)
		}

		logger.Logger.Info("Not found user")
		response.NotFound(w, "User not found")
		return
	}
	logger.Logger.Info("Found User: ", u.AccountName)

	roles, err := authorize.GetAllUserRolesMappingById(u.AccountID)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cannot save command to db: ", err)
		}

		logger.Logger.Error("Error to get role of user from db: ", err)
		response.InternalError(w, "Error to get role of user from db")
		return
	}
	logger.Logger.Info("User roles: ", roles)

	for _, role := range roles {
		if role.Permission != userDeleteReq.Permission {
			continue
		}
		err = authorize.DeleteUserRole(&db_models.CliRoleUserMapping{
			UserID:     u.AccountID,
			Permission: userDeleteReq.Permission,
		})
		if err != nil {
			loggerOperationHistory.ExecutedTime = time.Now()
			loggerOperationHistory.Result = "failure"
			err1 := history_command.SaveHistoryCommand(loggerOperationHistory)
			if err1 != nil {
				logger.Logger.Error("Cannot save command to db: ", err1)
			}

			logger.Logger.Error("Cannot Delete user role: ", err)
			response.InternalError(w, "Cannot Delete User Role")
			return
		}
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "success"
		err := history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cannot save command to db: ", err)
		}
		logger.Logger.Info("Delete User Role Success")
		response.Write(w, http.StatusOK, "Delete Success")
		return
	}

	loggerOperationHistory.ExecutedTime = time.Now()
	loggerOperationHistory.Result = "failure"
	err = history_command.SaveHistoryCommand(loggerOperationHistory)
	if err != nil {
		logger.Logger.Error("Cannot save command to db: ", err)
	}
	logger.Logger.Info("Cannot find permission of user")
	response.Write(w, http.StatusNotModified, "Cannot find permission of user")
	return
}
