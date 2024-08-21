package authorize

import (
	"fmt"
	"net/http"
	"serverGoChi/models/db_models"
	"serverGoChi/src/log"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authorize"
	"serverGoChi/src/service/history_command"
	"time"
)

func HandlerNeShow(w http.ResponseWriter, r *http.Request) {
	log.Logger.Info("Handler request authorize ne show")
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
		CmdName:     fmt.Sprintf("authorize ne show"),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}

	cliNeList, err := authorize.GetNeListBySystemType("5GC")
	if err != nil {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}

		log.Logger.Error("Cannot get cli ne list, err: ", err)
		response.InternalError(w, "Cannot get cli ne list")
		return
	}

	if cliNeList == nil || len(cliNeList) == 0 {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}

		log.Logger.Info("Empty NE List")
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

	logOperationHistory.ExecutedTime = time.Now()
	logOperationHistory.Result = "success"
	err = history_command.SaveHistoryCommand(logOperationHistory)
	if err != nil {
		log.Logger.Error("Cant save command to db: ", err)
	}
	response.Write(w, http.StatusFound, neShowRespList)
}
