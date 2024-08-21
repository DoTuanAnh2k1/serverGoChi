package authorize

import (
	"encoding/json"
	"fmt"
	"net/http"
	"serverGoChi/models/db_models"
	"serverGoChi/src/log"
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
	log.Logger.Info("Handler request authorize ne set")
	if r.Method != http.MethodPost {
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
		CmdName:     fmt.Sprintf("authorize ne show"),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}

	var neSetReq NeSetReq
	err := json.NewDecoder(r.Body).Decode(&neSetReq)
	if err != nil {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}

		log.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}

	log.Logger.Infof("Set ne with username: %v, ne: %v", neSetReq.Username, neSetReq.NeId)
	neId, err := strconv.ParseInt(neSetReq.NeId, 10, 64)
	if err != nil {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}

		log.Logger.Error("Error parsing integer: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing integer")
		return
	}

	tblAccount, err := user.GetUserByUserName(neSetReq.Username)
	if err != nil {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}

		log.Logger.Info("Cannot get user by username from db: ", err)
		response.Write(w, http.StatusInternalServerError, "Cant get user by username from db")
		return
	}

	cliNe, err := authorize.GetNeByNeId(neId)
	if err != nil {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}

		log.Logger.Info("Cannot get ne by ne id from db: ", err)
		response.Write(w, http.StatusInternalServerError, "Cannot get ne by ne id from db")
		return
	}

	if tblAccount == nil || cliNe == nil {
		log.Logger.Info("User or Ne Not Found")
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}
		response.NotFound(w, "User or Ne Not Found")
		return
	}

	neList, err := authenticate.GetNeListById(tblAccount.AccountID)
	if err != nil {
		log.Logger.Error("Cannot get ne list")
	}

	for _, ne := range neList {
		if ne.ID == neId {
			logOperationHistory.ExecutedTime = time.Now()
			logOperationHistory.Result = "failure"
			err = history_command.SaveHistoryCommand(logOperationHistory)
			if err != nil {
				log.Logger.Error("Cant save command to db: ", err)
			}

			log.Logger.Info("NeId already Assigned")
			response.Write(w, http.StatusNotModified, "NeId already Assigned")
			return
		}
	}

	err = authorize.AddUserCliNe(&db_models.CliUserNeMapping{
		UserID:  tblAccount.AccountID,
		TblNeID: neId,
	})
	if err != nil {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}

		log.Logger.Info("Cannot add cli ne to user: ", err)
		response.InternalError(w, "Cannot add cli ne to user")
		return
	}

	logOperationHistory.ExecutedTime = time.Now()
	logOperationHistory.Result = "success"
	err = history_command.SaveHistoryCommand(logOperationHistory)
	if err != nil {
		log.Logger.Error("Cant save command to db: ", err)
	}

	log.Logger.Info("Add cli ne to user")
	response.Write(w, http.StatusOK, "Add cli ne to user")
}
