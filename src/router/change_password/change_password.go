package change_password

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

func HandlerChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var userChangePasswordReq UserChangePasswordReq
	err := json.NewDecoder(r.Body).Decode(&userChangePasswordReq)
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
		CmdName:     fmt.Sprintf("change password for user %v", userMiddleware.Username),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}

	logger.Logger.Info("Handler request change password for user: ", userChangePasswordReq.Username)

	if userMiddleware.Username != userChangePasswordReq.Username {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}

		logger.Logger.Errorf("User %v do not have permission to change password user %v", userMiddleware.Username, userChangePasswordReq.Username)
		response.Unauthorized(w)
		return
	}

	u, err := user.GetUserByUserName(userChangePasswordReq.Username)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}

		logger.Logger.Info("Cant get user by username from db: ", err)
		response.Write(w, http.StatusInternalServerError, "Cant get user by username from db")
		return
	}

	if !bcrypt.Matches(userChangePasswordReq.Username+userChangePasswordReq.OldPassword, u.Password) {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}

		logger.Logger.Info("Wrong password")
		response.Write(w, http.StatusForbidden, "Wrong password")
		return
	}

	u.Password = bcrypt.Encode(userChangePasswordReq.Username + userChangePasswordReq.NewPassword)
	u.LockedTime = time.Now()
	err = user.UpdateUser(u)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}

		logger.Logger.Error("Cannot update password user: ", err)
		response.Write(w, http.StatusInternalServerError, "Cannot update password user")
		return
	}

	loggerOperationHistory.ExecutedTime = time.Now()
	loggerOperationHistory.Result = "success"
	err = history_command.SaveHistoryCommand(loggerOperationHistory)
	if err != nil {
		logger.Logger.Error("Cant save command to db: ", err)
	}
	logger.Logger.Info("Change password user success for user ", u.AccountName)
}

type UserChangePasswordReq struct {
	Username    string `json:"username"`
	NewPassword string `json:"new_password"`
	OldPassword string `json:"old_password"`
}
