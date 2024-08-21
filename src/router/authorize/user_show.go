package authorize

import (
	"fmt"
	"net/http"
	"serverGoChi/models/db_models"
	"serverGoChi/src/logger"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authenticate"
	"serverGoChi/src/service/history_command"
	"serverGoChi/src/service/user"
	"time"
)

func HandlerUserShow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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
		CmdName:     fmt.Sprintf("authorize-permission show"),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}
	logger.Logger.Info("Handler authorize user show")

	tblAccountList, err := user.GetAllUser()
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err1 := history_command.SaveHistoryCommand(loggerOperationHistory)
		if err1 != nil {
			logger.Logger.Error("Cannot save command to db: ", err1)
		}

		logger.Logger.Error("Get all tbl account from db fail: ", err)
		response.InternalError(w, "Get all tbl account from db fail")
		return
	}

	var userShowRespList []UserShowResp
	for _, tblAccount := range tblAccountList {
		roles, err := authenticate.GetRolesById(tblAccount.AccountID)
		if err != nil {
			logger.Logger.Error("Cannot get role, err: ", err)
			roles = ""
		}
		userShowRespList = append(userShowRespList, UserShowResp{
			Username:    tblAccount.AccountName,
			Permissions: roles,
		})
	}

	loggerOperationHistory.ExecutedTime = time.Now()
	loggerOperationHistory.Result = "success"
	err = history_command.SaveHistoryCommand(loggerOperationHistory)
	if err != nil {
		logger.Logger.Error("Cannot save command to db: ", err)
	}

	response.Write(w, http.StatusFound, userShowRespList)
}
