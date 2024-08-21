package list

import (
	"net/http"
	"serverGoChi/src/log"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authenticate"
	"serverGoChi/src/service/user"
)

func HandlerListNe(w http.ResponseWriter, r *http.Request) {
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

	var neResp NeResponse
	var neDataList []NeData
	for _, cliNe := range cliNeList {
		neDataList = append(neDataList, NeData{
			Site:        cliNe.SiteName,
			Ne:          cliNe.Name,
			Ip:          cliNe.IPAddress,
			Description: cliNe.Description,
			Namespace:   cliNe.Namespace,
			Port:        cliNe.Port,
			UrlList:     nil, // Need table tbl_ne to get this list
		})
	}

	if len(neDataList) == 0 {
		neResp.Code = "400"
		neResp.Status = "Fail"
		neResp.Message = "cannot find any ne belongs to the user"

		log.Logger.Info("cannot find any ne belongs to the user")
		response.Write(w, http.StatusNotFound, neResp)
		return
	}

	neResp.Code = "200"
	neResp.Status = "Found"
	neResp.Message = "Success"
	neResp.NeDataList = neDataList

	log.Logger.Info("Success")
	response.Write(w, http.StatusFound, neResp)
	return
}

type NeResponse struct {
	Status     string   `json:"status"`
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	NeDataList []NeData `json:"neDataList"`
}

type NeData struct {
	Site        string `json:"site"`
	Ne          string `json:"ne"`
	Ip          string `json:"ip"`
	Description string `json:"description"`
	Namespace   string `json:"namespace"`
	Port        int32  `json:"port"`
	UrlList     []Url  `json:"urlList"`
}

type Url struct {
	IpAddress string `json:"ipAddress"`
	Port      int    `json:"port"`
}
