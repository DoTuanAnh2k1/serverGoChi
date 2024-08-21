package authorize

import (
	"fmt"
	"net/http"
	"serverGoChi/models/db_models"
	"serverGoChi/src/log"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authenticate"
	"serverGoChi/src/service/history_command"
	"serverGoChi/src/service/user"
	"time"
)

func HandlerUserShow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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
		CmdName:     fmt.Sprintf("authorize-permission show"),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}
	log.Logger.Info("Handler authorize user show")

	tblAccountList, err := user.GetAllUser()
	if err != nil {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err1 := history_command.SaveHistoryCommand(logOperationHistory)
		if err1 != nil {
			log.Logger.Error("Cannot save command to db: ", err1)
		}

		log.Logger.Error("Get all tbl account from db fail: ", err)
		response.InternalError(w, "Get all tbl account from db fail")
		return
	}

	var userShowRespList []UserShowResp
	for _, tblAccount := range tblAccountList {
		roles, err := authenticate.GetRolesById(tblAccount.AccountID)
		if err != nil {
			log.Logger.Error("Cannot get role, err: ", err)
			roles = ""
		}
		userShowRespList = append(userShowRespList, UserShowResp{
			Username:    tblAccount.AccountName,
			Permissions: roles,
		})
	}

	logOperationHistory.ExecutedTime = time.Now()
	logOperationHistory.Result = "success"
	err = history_command.SaveHistoryCommand(logOperationHistory)
	if err != nil {
		log.Logger.Error("Cannot save command to db: ", err)
	}

	response.Write(w, http.StatusFound, userShowRespList)
}
