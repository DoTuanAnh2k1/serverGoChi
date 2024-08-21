package authorize

import (
	"encoding/json"
	"fmt"
	"net/http"
	"serverGoChi/models/db_models"
	"serverGoChi/src/logger"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authenticate"
	"serverGoChi/src/service/authorize"
	"serverGoChi/src/service/history_command"
	"serverGoChi/src/service/user"
	"strconv"
	"time"
)

func HandlerNeDelete(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Handler request authorize ne set")
	if r.Method != http.MethodPost {
		logger.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userMiddleware, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		logger.Logger.Error("Error to get user from token key")
		response.InternalError(w, "Internal Server Error")
		return
	}

	loggerOperationHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authorize ne show"),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}

	var neDeleteReq NeDeleteReq
	err := json.NewDecoder(r.Body).Decode(&neDeleteReq)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}

		logger.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}

	logger.Logger.Infof("Delete ne with username: %v, ne: %v", neDeleteReq.Username, neDeleteReq.NeId)
	neId, err := strconv.ParseInt(neDeleteReq.NeId, 10, 64)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}

		logger.Logger.Error("Error parsing integer: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing integer")
		return
	}

	tblAccount, err := user.GetUserByUserName(neDeleteReq.Username)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}

		logger.Logger.Info("Cannot get user by username from db: ", err)
		response.Write(w, http.StatusInternalServerError, "Cant get user by username from db")
		return
	}

	cliNe, err := authorize.GetNeByNeId(neId)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}

		logger.Logger.Info("Cannot get ne by ne id from db: ", err)
		response.Write(w, http.StatusInternalServerError, "Cannot get ne by ne id from db")
		return
	}

	if tblAccount == nil || cliNe == nil {
		logger.Logger.Info("User or Ne Not Found")
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		response.NotFound(w, "User or Ne Not Found")
		return
	}

	neList, err := authenticate.GetNeListById(tblAccount.AccountID)
	if err != nil {
		logger.Logger.Error("Cannot get ne list")
	}

	for _, ne := range neList {
		if ne.ID != neId {
			continue
		}

		err = authorize.DeleteCliNe(&db_models.CliUserNeMapping{
			UserID:  tblAccount.AccountID,
			TblNeID: neId,
		})
		if err != nil {
			loggerOperationHistory.ExecutedTime = time.Now()
			loggerOperationHistory.Result = "failure"
			err = history_command.SaveHistoryCommand(loggerOperationHistory)
			if err != nil {
				logger.Logger.Error("Cant save command to db: ", err)
			}

			logger.Logger.Info("Cannot delete cli ne to user: ", err)
			response.InternalError(w, "Cannot delete cli ne to user")
			return
		}

		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "success"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}

		logger.Logger.Info("Delete cli ne to user")
		response.Write(w, http.StatusOK, "Delete cli ne to user")
		return
	}

	loggerOperationHistory.ExecutedTime = time.Now()
	loggerOperationHistory.Result = "failure"
	err = history_command.SaveHistoryCommand(loggerOperationHistory)
	if err != nil {
		logger.Logger.Error("Cant save command to db: ", err)
	}

	logger.Logger.Info("Not found user and ne relationship")
	response.Write(w, http.StatusNotModified, "Not found user and ne relationship")
	return
}
