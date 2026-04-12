package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// HandlerNeShow handles GET /aa/authorize/ne/show
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
		Scope:       "cli-config",
		Account:     userMiddleware.Username,
	}

	cliNeList, err := service.GetNeListBySystemType("5GC")
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = service.SaveHistoryCommand(loggerOperationHistory)
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
		err = service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		logger.Logger.Info("Empty NE List")
		response.NotFound(w, "Empty NE List")
		return
	}

	var neShowRespList []neShowResp
	for _, cliNe := range cliNeList {
		neShowRespList = append(neShowRespList, neShowResp{
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
	err = service.SaveHistoryCommand(loggerOperationHistory)
	if err != nil {
		logger.Logger.Error("Cant save command to db: ", err)
	}
	response.Write(w, http.StatusFound, neShowRespList)
}

// HandlerNeRemove handles POST /aa/authorize/ne/remove
func HandlerNeRemove(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == 0 {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}

	userMiddleware, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		response.InternalError(w, "Internal Server Error")
		return
	}

	opHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authorize ne remove id %v", req.ID),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     userMiddleware.Username,
	}

	if err := service.DeleteNeById(req.ID); err != nil {
		logger.Logger.Error("authorize/ne/remove: ", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to delete NE")
		return
	}

	saveHistory(opHistory, "success")
	response.Success(w, "NE deleted")
}

// HandlerNeCreate handles POST /aa/authorize/ne/create
func HandlerNeCreate(w http.ResponseWriter, r *http.Request) {
	var req db_models.CliNe
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Error("authorize/ne/create: decode request body: ", err)
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		response.Write(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.SystemType == "" {
		req.SystemType = "5GC"
	}

	userMiddleware, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		response.InternalError(w, "Internal Server Error")
		return
	}

	opHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authorize ne create name %v", req.Name),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     userMiddleware.Username,
	}

	if err := service.CreateNe(&req); err != nil {
		logger.Logger.Error("authorize/ne/create: ", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to create NE")
		return
	}

	saveHistory(opHistory, "success")
	response.Created(w)
}

// HandlerNeSet handles POST /aa/authorize/ne/set
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
		CmdName:     fmt.Sprintf("authorize ne set"),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     userMiddleware.Username,
	}

	var req neSetReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		logger.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}

	logger.Logger.Infof("Set ne with username: %v, ne: %v", req.Username, req.NeId)
	neId, err := strconv.ParseInt(req.NeId, 10, 64)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		logger.Logger.Error("Error parsing integer: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing integer")
		return
	}

	tblAccount, err := service.GetUserByUserName(req.Username)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		logger.Logger.Info("Cannot get user by username from db: ", err)
		response.Write(w, http.StatusInternalServerError, "Cant get user by username from db")
		return
	}

	cliNe, err := service.GetNeByNeId(neId)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = service.SaveHistoryCommand(loggerOperationHistory)
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
		err = service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		response.NotFound(w, "User or Ne Not Found")
		return
	}

	neList, err := service.GetAllCliNeOfUserByUserId(tblAccount.AccountID)
	if err != nil {
		logger.Logger.Error("Cannot get ne list")
	}

	for _, ne := range neList {
		if ne.TblNeID == neId {
			loggerOperationHistory.ExecutedTime = time.Now()
			loggerOperationHistory.Result = "failure"
			err = service.SaveHistoryCommand(loggerOperationHistory)
			if err != nil {
				logger.Logger.Error("Cant save command to db: ", err)
			}
			logger.Logger.Info("NeId already Assigned")
			response.Write(w, http.StatusNotModified, "NeId already Assigned")
			return
		}
	}

	err = service.AddUserCliNe(&db_models.CliUserNeMapping{
		UserID:  tblAccount.AccountID,
		TblNeID: neId,
	})
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		logger.Logger.Info("Cannot add cli ne to user: ", err)
		response.InternalError(w, "Cannot add cli ne to user")
		return
	}

	loggerOperationHistory.ExecutedTime = time.Now()
	loggerOperationHistory.Result = "success"
	err = service.SaveHistoryCommand(loggerOperationHistory)
	if err != nil {
		logger.Logger.Error("Cant save command to db: ", err)
	}
	logger.Logger.Info("Add cli ne to user")
	response.Write(w, http.StatusOK, "Add cli ne to user")
}

// HandlerNeDelete handles POST /aa/authorize/ne/delete
func HandlerNeDelete(w http.ResponseWriter, r *http.Request) {
	logger.Logger.Info("Handler request authorize ne delete")
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
		CmdName:     fmt.Sprintf("authorize ne delete"),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     userMiddleware.Username,
	}

	var req neDeleteReq
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		logger.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}

	logger.Logger.Infof("Delete ne with username: %v, ne: %v", req.Username, req.NeId)
	neId, err := strconv.ParseInt(req.NeId, 10, 64)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		logger.Logger.Error("Error parsing integer: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing integer")
		return
	}

	tblAccount, err := service.GetUserByUserName(req.Username)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		logger.Logger.Info("Cannot get user by username from db: ", err)
		response.Write(w, http.StatusInternalServerError, "Cant get user by username from db")
		return
	}

	cliNe, err := service.GetNeByNeId(neId)
	if err != nil {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err = service.SaveHistoryCommand(loggerOperationHistory)
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
		err = service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		response.NotFound(w, "User or Ne Not Found")
		return
	}

	neList, err := service.GetAllCliNeOfUserByUserId(tblAccount.AccountID)
	if err != nil {
		logger.Logger.Error("Cannot get ne list")
	}

	for _, ne := range neList {
		if ne.TblNeID != neId {
			continue
		}

		err = service.DeleteCliNe(&db_models.CliUserNeMapping{
			UserID:  tblAccount.AccountID,
			TblNeID: neId,
		})
		if err != nil {
			loggerOperationHistory.ExecutedTime = time.Now()
			loggerOperationHistory.Result = "failure"
			err = service.SaveHistoryCommand(loggerOperationHistory)
			if err != nil {
				logger.Logger.Error("Cant save command to db: ", err)
			}
			logger.Logger.Info("Cannot delete cli ne to user: ", err)
			response.InternalError(w, "Cannot delete cli ne to user")
			return
		}

		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "success"
		err = service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		logger.Logger.Info("Delete cli ne to user")
		response.Write(w, http.StatusOK, "Delete cli ne to user")
		return
	}

	loggerOperationHistory.ExecutedTime = time.Now()
	loggerOperationHistory.Result = "failure"
	err = service.SaveHistoryCommand(loggerOperationHistory)
	if err != nil {
		logger.Logger.Error("Cant save command to db: ", err)
	}
	logger.Logger.Info("Not found user and ne relationship")
	response.Write(w, http.StatusNotModified, "Not found user and ne relationship")
}

// HandlerListNe handles GET /aa/list/ne
func HandlerListNe(w http.ResponseWriter, r *http.Request) {
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
	logger.Logger.Info("Handler list ne")

	tblAccount, err := service.GetUserByUserName(userMiddleware.Username)
	if err != nil {
		logger.Logger.Error("Cannot get user by username from db: ", err)
		response.InternalError(w, "Cannot get user by username from db")
		return
	}

	cliUserNeMappingList, err := service.GetAllCliNeOfUserByUserId(tblAccount.AccountID)
	if err != nil {
		logger.Logger.Error("Cannot list cli user ne mapping from db: ", err)
		response.InternalError(w, "Cannot list cli user ne mapping from db")
		return
	}

	var neResp neResponse
	var neDataList []neData
	for _, cliUserNeMapping := range cliUserNeMappingList {
		cliNe, err := service.GetNeByNeId(cliUserNeMapping.TblNeID)
		if err != nil {
			logger.Logger.Error("Cannot list cli ne from db: ", err)
			response.InternalError(w, "Cannot list cli ne from db")
			return
		}
		neDataList = append(neDataList, neData{
			Site:        cliNe.SiteName,
			Ne:          cliNe.Name,
			Ip:          cliNe.IPAddress,
			Description: cliNe.Description,
			Namespace:   cliNe.Namespace,
			Port:        cliNe.Port,
			UrlList:     nil,
		})
	}

	if len(neDataList) == 0 {
		neResp.Code = "400"
		neResp.Status = "Fail"
		neResp.Message = "cannot find any ne belongs to the user"
		logger.Logger.Info("cannot find any ne belongs to the user")
		response.Write(w, http.StatusNotFound, neResp)
		return
	}

	neResp.Code = "200"
	neResp.Status = "Found"
	neResp.Message = "Success"
	neResp.NeDataList = neDataList
	logger.Logger.Info("Success")
	response.Write(w, http.StatusFound, neResp)
}

// HandlerListNeMonitor handles GET /aa/list/ne/monitor
func HandlerListNeMonitor(w http.ResponseWriter, r *http.Request) {
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
	logger.Logger.Info("Handler list ne monitor")

	tblAccount, err := service.GetUserByUserName(userMiddleware.Username)
	if err != nil {
		logger.Logger.Error("Cannot get user by username from db: ", err)
		response.InternalError(w, "Cannot get user by username from db")
		return
	}

	cliUserNeMappingList, err := service.GetAllCliNeOfUserByUserId(tblAccount.AccountID)
	if err != nil {
		logger.Logger.Error("Cannot list cli user ne mapping from db: ", err)
		response.InternalError(w, "Cannot list cli user ne mapping from db")
		return
	}

	if len(cliUserNeMappingList) == 0 {
		logger.Logger.Info("Cli Ne List is empty")
		response.Write(w, http.StatusNotFound, "Cli Ne List is empty")
		return
	}

	var neMonitorInfoList []neMonitorInfo
	for _, cliUserNeMapping := range cliUserNeMappingList {
		cliNe, err := service.GetNeByNeId(cliUserNeMapping.TblNeID)
		if err != nil {
			logger.Logger.Error("Cannot list cli ne from db: ", err)
			response.InternalError(w, "Cannot list cli ne from db")
			return
		}
		neMonitor, err := service.GetNeMonitorById(cliNe.ID)
		if err != nil || neMonitor == nil {
			logger.Logger.Error("Ne Monitor is null")
			continue
		}
		neMonitorInfoList = append(neMonitorInfoList, neMonitorInfo{
			Site:         cliNe.SiteName,
			Ne:           cliNe.Name,
			Ip:           cliNe.IPAddress,
			Description:  cliNe.Description,
			Namespace:    cliNe.Namespace,
			Port:         cliNe.Port,
			NeMonitorURL: neMonitor.NeIP,
		})
	}

	logger.Logger.Info("End of handler request list ne monitor")
	response.Write(w, http.StatusOK, neMonitorInfoList)
}

// NE request/response types
type neSetReq struct {
	Username string `json:"username"`
	NeId     string `json:"neid"`
}

type neDeleteReq struct {
	Username string `json:"username"`
	NeId     string `json:"neid"`
}

type neShowResp struct {
	Name        string `json:"name"`
	SiteName    string `json:"site_name"`
	IpAddress   string `json:"ip_address"`
	Port        int32  `json:"port"`
	Description string `json:"description"`
	Id          int64  `json:"id"`
}

type neResponse struct {
	Status     string   `json:"status"`
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	NeDataList []neData `json:"neDataList"`
}

type neData struct {
	Site        string  `json:"site"`
	Ne          string  `json:"ne"`
	Ip          string  `json:"ip"`
	Description string  `json:"description"`
	Namespace   string  `json:"namespace"`
	Port        int32   `json:"port"`
	UrlList     []neURL `json:"urlList"`
}

type neURL struct {
	IpAddress string `json:"ipAddress"`
	Port      int    `json:"port"`
}

type neMonitorInfo struct {
	Site         string `json:"site"`
	Ne           string `json:"ne"`
	Ip           string `json:"ip"`
	Description  string `json:"description"`
	Namespace    string `json:"namespace"`
	Port         int32  `json:"port"`
	NeMonitorURL string `json:"ne_monitor_url"`
}
