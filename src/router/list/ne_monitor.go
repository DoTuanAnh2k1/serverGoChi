package list

import (
	"net/http"
	"serverGoChi/src/log"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authenticate"
	"serverGoChi/src/service/list"
	"serverGoChi/src/service/user"
)

func HandlerListNeMonitor(w http.ResponseWriter, r *http.Request) {
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
	log.Logger.Info("Handler list ne")

	tblAccount, err := user.GetUserByUserName(userMiddleware.Username)
	if err != nil {
		log.Logger.Error("Cannot get user by username from db: ", err)
		response.InternalError(w, "Cannot get user by username from db")
		return
	}

	cliNeList, err := authenticate.GetNeListById(tblAccount.AccountID)
	if err != nil {
		log.Logger.Error("Cannot list cli ne from db: ", err)
		response.InternalError(w, "Cannot list cli ne from db")
		return
	}

	if len(cliNeList) == 0 {
		log.Logger.Info("Cli Ne List is empty")
		response.Write(w, http.StatusNotFound, "Cli Ne List is empty")
		return
	}

	var neMonitorInfoList []NeMonitorInfo
	for _, cliNe := range cliNeList {
		neMonitor, err := list.GetNeMonitorById(cliNe.ID)
		if err != nil || neMonitor == nil {
			log.Logger.Error("Ne Monitor is null")
			continue
		}
		neMonitorInfoList = append(neMonitorInfoList, NeMonitorInfo{
			Site:         cliNe.SiteName,
			Ne:           cliNe.Name,
			Ip:           cliNe.IPAddress,
			Description:  cliNe.Description,
			Namespace:    cliNe.Namespace,
			Port:         cliNe.Port,
			NeMonitorURL: neMonitor.NeIP,
		})
	}

	log.Logger.Info("End of handler request list ne monitor")
	response.Write(w, http.StatusOK, "Success")
	return
}
