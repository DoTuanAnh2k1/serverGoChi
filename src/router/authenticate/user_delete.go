package authenticate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"serverGoChi/models/db_models"
	"serverGoChi/src/log"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/history_command"
	"serverGoChi/src/service/user"
	"time"
)

func HandlerAuthenticateUserDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var userInfo db_models.TblAccount
	err := json.NewDecoder(r.Body).Decode(&userInfo)
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
		CmdName:     fmt.Sprintf("authenticate-user delete username %v password xxx", userInfo.AccountName),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}

	u, err := user.GetUserByUserName(userInfo.AccountName)
	if err != nil {
		log.Logger.Error("Cant get user by username from db: ", err)
		response.Write(w, http.StatusInternalServerError, "Cant get user by username from db")
		return
	}

	if u != nil && u.IsEnable == true {
		u.IsEnable = false
		err := user.AddUser(*u)
		if err != nil {
			log.Logger.Error("Cant update user to db: ", err)
			response.Write(w, http.StatusInternalServerError, "Cant update user to db")
			return
		}

		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "success"
		err = history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}

		response.Success(w, "")
	} else {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}

		response.Write(w, http.StatusNotFound, "User Not Found")
	}
}
