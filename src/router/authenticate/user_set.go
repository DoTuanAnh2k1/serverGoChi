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
	"serverGoChi/src/utils/bcrypt"
	"time"
)

func HandlerAuthenticateUserSet(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Handler Authenticate set user")
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
		logger.Logger.Error("Error to get user from context")
		response.InternalError(w, "Internal Server Error")
		return
	}

	loggerOperationHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authenticate-user set username %v password xxx", userInfo.AccountName),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}

	u, err := user.GetUserByUserName(userInfo.AccountName)
	if err != nil {
		logger.Logger.Info("Cant get user by username from db: ", err)
		// response.Write(w, http.StatusInternalServerError, "Cant get user by username from db")
		// return
	}

	hashPass := bcrypt.Encode(userInfo.AccountName + userInfo.Password)
	userInfo.Password = hashPass

	if u == nil {
		userInfo.IsEnable = true
		userInfo.CreatedBy = userMiddleware.Username
		userInfo.CreatedDate = time.Now()
		userInfo.UpdatedDate = time.Now()
		userInfo.LastLoginTime = time.Now()
		userInfo.LastChangePass = time.Now()
		userInfo.LockedTime = time.Now()
		userInfo.AccountType = 2
		userInfo.Status = true
		err := user.AddUser(&userInfo)
		if err != nil {
			logger.Logger.Error("Cant create user to db: ", err)
			response.Write(w, http.StatusInternalServerError, "Cant create user to db")
			return
		}

		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "success"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		logger.Logger.Info("Create user")
		response.Created(w)
	}

	if !u.IsEnable {
		u.IsEnable = true
		u.CreatedBy = userMiddleware.Username
		u.UpdatedDate = time.Now()
		userInfo.LoginFailureCount = 0
		err := user.UpdateUser(u)
		if err != nil {
			logger.Logger.Error("Cant create user to db: ", err)
			response.Write(w, http.StatusInternalServerError, "Cant get user by username from db")
			return
		}

		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "success"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		logger.Logger.Info("Enable user with username: ", u.AccountName)
		response.Created(w)
	} else {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		logger.Logger.Info("User already exist, nothing change")
		response.Write(w, http.StatusNotModified, "User already exist!")
	}
}
