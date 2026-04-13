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

// HandlerNeShow liệt kê tất cả NE thuộc hệ thống 5GC.
//
// Input : GET (không có body/query params)
// Output: 302 [ { name, site_name, ip_address, port, description, id } ]
//         404 nếu danh sách NE rỗng
//         500 nếu lỗi DB
// Flow  : lấy actor từ context → GetNeListBySystemType("5GC") →
//         map sang neShowResp → ghi operation history → trả danh sách
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
			Namespace:   cliNe.Namespace,
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

// HandlerNeUpdate cập nhật thông tin một NE (site, IP, port, namespace, description).
//
// Input : POST body JSON { "id": int64 (bắt buộc), cùng các trường CliNe muốn cập nhật }
// Output: 200 "NE updated" nếu thành công
//         400 nếu thiếu id hoặc body không hợp lệ
//         500 nếu lỗi DB
// Flow  : decode body → validate id > 0 → lấy actor từ context →
//         UpdateNe → ghi operation history
func HandlerNeUpdate(w http.ResponseWriter, r *http.Request) {
	var req db_models.CliNe
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == 0 {
		response.Write(w, http.StatusBadRequest, "id is required")
		return
	}

	user := mustUser(r)
	op := opHistory("ne update", fmt.Sprintf("id=%d name=%s", req.ID, req.Name), user.Username)

	if err := service.UpdateNe(&req); err != nil {
		logger.Logger.Error("authorize/ne/update: ", err)
		saveHistory(op, "failure")
		response.InternalError(w, "failed to update NE")
		return
	}

	saveHistory(op, "success")
	response.Success(w, "NE updated")
}

// HandlerNeRemove xoá một NE và toàn bộ dữ liệu liên quan (cascade).
//
// Input : POST body JSON { "id": int64 }
// Output: 200 "NE deleted" nếu thành công
//         400 nếu body không hợp lệ hoặc id == 0
//         500 nếu lỗi DB khi xoá cascade
// Flow  : decode body → kiểm tra id > 0 → lấy actor từ context →
//         DeleteNeById (cascade: user_ne_mapping → ne_monitor → ne_config → ne_slave → ne) →
//         ghi operation history
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

// HandlerNeCreate tạo mới một NE trong hệ thống.
//
// Input : POST body JSON { "name": string (bắt buộc), "site_name", "ip_address",
//         "port", "namespace", "description", "system_type" } — các trường CliNe
// Output: 201 nếu tạo thành công
//         400 nếu thiếu name hoặc body không hợp lệ
//         500 nếu lỗi DB
// Flow  : decode body → validate name không rỗng →
//         mặc định system_type="5GC" nếu không có → lấy actor từ context →
//         CreateNe → ghi operation history
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

// HandlerNeSet gán một NE cho user (tạo user-NE mapping).
//
// Input : POST body JSON { "username": string, "neid": string (ID dạng số) }
// Output: 200 "Add cli ne to user" nếu thành công
//         304 nếu NE đã được gán cho user này rồi
//         404 nếu user hoặc NE không tồn tại
//         500 nếu lỗi parse/DB
// Flow  : decode body → parse neId sang int64 → lấy actor từ context →
//         GetUserByUserName → GetNeByNeId → kiểm tra mapping chưa tồn tại →
//         AddUserCliNe → ghi operation history
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

// HandlerNeDelete gỡ bỏ mapping giữa user và một NE.
//
// Input : POST body JSON { "username": string, "neid": string (ID dạng số) }
// Output: 200 "Delete cli ne to user" nếu thành công
//         304 nếu mapping không tồn tại
//         404 nếu user hoặc NE không tồn tại
//         500 nếu lỗi parse/DB
// Flow  : decode body → parse neId sang int64 → lấy actor từ context →
//         GetUserByUserName → GetNeByNeId → tìm mapping tương ứng →
//         DeleteCliNe → ghi operation history
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

// HandlerListNe trả về danh sách NE mà user hiện tại được phép truy cập.
//
// Input : GET (không có body/query params; user lấy từ JWT context)
// Output: 302 { status, code, message, neDataList: [{site, ne, ip, description, namespace, port}] }
//         404 nếu user không có NE nào được gán
//         500 nếu lỗi DB
// Flow  : lấy actor từ context → GetUserByUserName → GetAllCliNeOfUserByUserId →
//         với mỗi mapping lấy GetNeByNeId → gộp danh sách neData
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

// HandlerListNeMonitor trả về thông tin monitor URL của các NE thuộc user (mode command).
//
// Input : GET (không có body/query params; user lấy từ JWT context)
// Output: 200 [ { site, ne, ip, description, namespace, port, ne_monitor_url } ]
//         404 nếu user không có NE nào được gán
//         500 nếu lỗi DB
// Flow  : lấy actor từ context → GetUserByUserName → GetAllCliNeOfUserByUserId →
//         với mỗi NE: GetNeByNeId → GetNeMonitorById → lấy NeMonitor.NeIP làm URL →
//         gộp danh sách neMonitorInfo
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
	Namespace   string `json:"namespace"`
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
