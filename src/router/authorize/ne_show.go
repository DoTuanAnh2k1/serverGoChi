package authorize

import (
	"fmt"
	"net/http"
	"serverGoChi/models/db_models"
	"serverGoChi/src/logger"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authorize"
	"serverGoChi/src/service/history_command"
	"time"
)

func HandlerNeShow(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Handler request authorize ne show")
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
		CmdName:     fmt.Sprintf("authorize ne show"),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}

	cliNeList, err := authorize.GetNeListBySystemType("5GC")
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}

		logger.Logger.Error("Cannot get cli ne list, err: ", err)
		response.InternalError(w, "Cannot get cli ne list")
		return
	}

	if cliNeList == nil || len(cliNeList) == 0 {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}

		logger.Logger.Info("Empty NE List")
		response.NotFound(w, "Empty NE List")
		return
	}

	var neShowRespList []NeShowResp
	for _, cliNe := range cliNeList {
		neShowRespList = append(neShowRespList, NeShowResp{
			Name:        cliNe.Name,
			SiteName:    cliNe.SiteName,
			IpAddress:   cliNe.IPAddress,
			Port:        cliNe.Port,
			Description: cliNe.Description,
			Id:          cliNe.ID,
		})
	}

	loggerOperationHistory.ExecutedTime = time.Now()
	loggerOperationHistory.Result = "success"
	err = history_command.SaveHistoryCommand(loggerOperationHistory)
	if err != nil {
		logger.Logger.Error("Cant save command to db: ", err)
	}
	response.Write(w, http.StatusFound, neShowRespList)
}
