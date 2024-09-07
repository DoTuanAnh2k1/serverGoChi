package authenticate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"serverGoChi/models/db_models"
	"serverGoChi/src/logger"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/history_command"
	"serverGoChi/src/service/user"
	"time"
)

func HandlerAuthenticateUserDelete(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Handler Authenticate delete user")
	if r.Method != http.MethodPost {
		logger.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var userInfo db_models.TblAccount
	err := json.NewDecoder(r.Body).Decode(&userInfo)
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
		CmdName:     fmt.Sprintf("authenticate-user delete username %v password xxx", userInfo.AccountName),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}

	u, err := user.GetUserByUserName(userInfo.AccountName)
	if err != nil {
		logger.Logger.Error("Cant get user by username from db: ", err)
		response.Write(w, http.StatusInternalServerError, "Cant get user by username from db")
		return
	}

	if u != nil && u.IsEnable == true {
		u.IsEnable = false
		u.LockedTime = time.Now()
		err := user.UpdateUser(u)
		if err != nil {
			logger.Logger.Error("Cant update user to db: ", err)
			response.Write(w, http.StatusInternalServerError, "Cant update user to db")
			return
		}

		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "success"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		logger.Logger.Info("Disable user with username: ", u.AccountName)
		response.Success(w, "")
	} else {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		if u == nil {
			logger.Logger.Info("Not found user: ", u.AccountName)
		}
		if !u.IsEnable {
			logger.Logger.Info("User already disable")
		}
		response.Write(w, http.StatusNotFound, "User Not Found")
	}
}
