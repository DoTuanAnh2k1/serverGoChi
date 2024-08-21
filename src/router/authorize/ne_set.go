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

func HandlerNeSet(w http.ResponseWriter, r *http.Request) {
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

	var neSetReq NeSetReq
	err := json.NewDecoder(r.Body).Decode(&neSetReq)
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

	logger.Logger.Infof("Set ne with username: %v, ne: %v", neSetReq.Username, neSetReq.NeId)
	neId, err := strconv.ParseInt(neSetReq.NeId, 10, 64)
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

	tblAccount, err := user.GetUserByUserName(neSetReq.Username)
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
		if ne.ID == neId {
			loggerOperationHistory.ExecutedTime = time.Now()
			loggerOperationHistory.Result = "failure"
			err = history_command.SaveHistoryCommand(loggerOperationHistory)
			if err != nil {
				logger.Logger.Error("Cant save command to db: ", err)
			}

			logger.Logger.Info("NeId already Assigned")
			response.Write(w, http.StatusNotModified, "NeId already Assigned")
			return
		}
	}

	err = authorize.AddUserCliNe(&db_models.CliUserNeMapping{
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

		logger.Logger.Info("Cannot add cli ne to user: ", err)
		response.InternalError(w, "Cannot add cli ne to user")
		return
	}

	loggerOperationHistory.ExecutedTime = time.Now()
	loggerOperationHistory.Result = "success"
	err = history_command.SaveHistoryCommand(loggerOperationHistory)
	if err != nil {
		logger.Logger.Error("Cant save command to db: ", err)
	}

	logger.Logger.Info("Add cli ne to user")
	response.Write(w, http.StatusOK, "Add cli ne to user")
}
