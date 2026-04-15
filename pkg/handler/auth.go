package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/bcrypt"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/token"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// HandlerAuthenticate xác thực user và cấp JWT token.
//
// Input : POST body JSON { "username": string, "password": string }
// Output: 200 { status, response_data (JWT), response_code, system_type }
//         400 nếu thiếu username/password hoặc body không hợp lệ
//         401 nếu sai thông tin đăng nhập hoặc DB lỗi khi verify
//         500 nếu không lấy được roles, tạo token thất bại, hoặc ghi login history lỗi
// Flow  : decode body → validate username/password không rỗng →
//         Authenticate (bcrypt check) → GetRolesById → CreateToken →
//         UpdateLoginHistory → trả JWT token
func HandlerAuthenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req authUser
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Errorf("authenticate: decode request body: %v", err)
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(req.UserName) == "" || strings.TrimSpace(req.Password) == "" {
		logger.Logger.Warn("authenticate: missing username or password")
		response.Write(w, http.StatusBadRequest, "username and password are required")
		return
	}

	log := logger.Logger.WithField("user", req.UserName).WithField("ip", r.RemoteAddr)

	ok, err, _ := service.Authenticate(req.UserName, req.Password)
	if err != nil || !ok {
		log.Warn("authenticate: login failed")
		response.Unauthorized(w)
		return
	}

	u, err := service.GetUserByUserName(req.UserName)
	if err != nil || u == nil {
		log.Errorf("authenticate: get user: %v", err)
		response.InternalError(w, "failed to load user")
		return
	}
	permission := service.GetPermissionByUser(u)

	tokenString, err := token.CreateToken(req.UserName, permission)
	if err != nil {
		log.Errorf("authenticate: create token: %v", err)
		response.InternalError(w, "failed to create token")
		return
	}

	if err = service.UpdateLoginHistory(req.UserName, r.RemoteAddr); err != nil {
		log.Errorf("authenticate: update login history: %v", err)
		response.InternalError(w, "failed to record login")
		return
	}

	log.Infof("authenticate: login successful — permission=%q", permission)
	response.Write(w, http.StatusOK, tokenRequestResponse{
		Status:       "success",
		ResponseData: tokenString,
		ResponseCode: "200",
		SystemType:   "5GC",
	})
}

// HandlerAuthenticateUserSet tạo mới hoặc kích hoạt lại một tài khoản.
//
// Input : POST body JSON { "account_name": string, "password": string, ...TblAccount fields }
// Output: 201 nếu tạo mới hoặc re-enable thành công
//         400 nếu thiếu username/password hoặc body không hợp lệ
//         304 nếu user đã tồn tại và đang active
//         500 nếu lỗi DB khi tạo hoặc cập nhật user
// Flow  : decode body → validate username/password → lấy actor từ context →
//         GetUserByUserName → nếu không tồn tại thì AddUser (hash password, set defaults) →
//         nếu đang bị disable thì re-enable → ghi operation history
func HandlerAuthenticateUserSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var userInfo db_models.TblAccount
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		logger.Logger.Errorf("authenticate/user/set: decode request body: %v", err)
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if strings.TrimSpace(userInfo.AccountName) == "" {
		response.Write(w, http.StatusBadRequest, "username is required")
		return
	}
	if strings.TrimSpace(userInfo.Password) == "" {
		response.Write(w, http.StatusBadRequest, "password is required")
		return
	}

	actor, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		logger.Logger.Error("authenticate/user/set: user not found in context")
		response.InternalError(w, "Internal Server Error")
		return
	}

	log := logger.Logger.WithField("actor", actor.Username).WithField("target", userInfo.AccountName)

	opHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authenticate-user set username %v password xxx", userInfo.AccountName),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     actor.Username,
	}

	existing, err := service.GetUserByUserName(userInfo.AccountName)
	if err != nil {
		log.Errorf("authenticate/user/set: get user: %v", err)
	}

	hashPass := bcrypt.Encode(userInfo.AccountName + userInfo.Password)
	userInfo.Password = hashPass

	if existing == nil {
		userInfo.IsEnable = true
		userInfo.CreatedBy = actor.Username
		now := time.Now()
		userInfo.CreatedDate, userInfo.UpdatedDate = now, now
		userInfo.LastLoginTime, userInfo.LastChangePass, userInfo.LockedTime = now, now, now
		userInfo.AccountType = 2
		userInfo.Status = true

		if err := service.AddUser(&userInfo); err != nil {
			log.Errorf("authenticate/user/set: create user: %v", err)
			saveHistory(opHistory, "failure")
			response.InternalError(w, "failed to create user")
			return
		}
		log.Info("authenticate/user/set: user created")
		saveHistory(opHistory, "success")
		response.Created(w)
		return
	}

	if !existing.IsEnable {
		existing.IsEnable = true
		existing.CreatedBy = actor.Username
		existing.UpdatedDate = time.Now()
		existing.LoginFailureCount = 0
		if err := service.UpdateUser(existing); err != nil {
			log.Errorf("authenticate/user/set: re-enable user: %v", err)
			saveHistory(opHistory, "failure")
			response.InternalError(w, "failed to enable user")
			return
		}
		log.Info("authenticate/user/set: user re-enabled")
		saveHistory(opHistory, "success")
		response.Created(w)
		return
	}

	log.Warn("authenticate/user/set: user already active — no change")
	saveHistory(opHistory, "failure")
	response.Write(w, http.StatusNotModified, "user already exists")
}

// HandlerAuthenticateUserDelete vô hiệu hoá (soft-delete) một tài khoản.
//
// Input : POST body JSON { "account_name": string }
// Output: 200 nếu disable thành công
//         404 nếu user không tồn tại hoặc đã bị disable
//         500 nếu lỗi DB khi tìm hoặc cập nhật user
// Flow  : decode body → lấy actor từ context → GetUserByUserName →
//         nếu user active thì set IsEnable=false, cập nhật LockedTime → UpdateUser →
//         ghi operation history
func HandlerAuthenticateUserDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var userInfo db_models.TblAccount
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		logger.Logger.Errorf("authenticate/user/delete: decode request body: %v", err)
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actor, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		logger.Logger.Error("authenticate/user/delete: user not found in context")
		response.InternalError(w, "Internal Server Error")
		return
	}

	log := logger.Logger.WithField("actor", actor.Username).WithField("target", userInfo.AccountName)

	opHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authenticate-user delete username %v", userInfo.AccountName),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     actor.Username,
	}

	u, err := service.GetUserByUserName(userInfo.AccountName)
	if err != nil {
		log.Errorf("authenticate/user/delete: get user: %v", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to find user")
		return
	}

	if u != nil && u.IsEnable {
		u.IsEnable = false
		u.LockedTime = time.Now()
		if err := service.UpdateUser(u); err != nil {
			log.Errorf("authenticate/user/delete: disable user: %v", err)
			saveHistory(opHistory, "failure")
			response.InternalError(w, "failed to disable user")
			return
		}
		log.Info("authenticate/user/delete: user disabled")
		saveHistory(opHistory, "success")
		response.Success(w, "")
		return
	}

	log.Warn("authenticate/user/delete: user not found or already disabled")
	saveHistory(opHistory, "failure")
	response.Write(w, http.StatusNotFound, "user not found")
}

// HandlerAuthenticateUserShow liệt kê tất cả user cùng NE và role tương ứng.
//
// Input : GET (không có body/query params)
// Output: 302 [ { username, tblNes: [{ne, site}], role } ]
//         404 nếu không có user nào
//         500 nếu lỗi DB
// Flow  : lấy actor từ context → GetAllUser → với mỗi user:
//         GetAllCliNeOfUserByUserId → GetNeByNeId (lấy tên NE/site) →
//         GetRolesById → gộp kết quả → ghi operation history
func HandlerAuthenticateUserShow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	actor, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		logger.Logger.Error("authenticate/user/show: user not found in context")
		response.InternalError(w, "Internal Server Error")
		return
	}

	opHistory := db_models.CliOperationHistory{
		CmdName:     "authenticate-user show",
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     actor.Username,
	}

	userList, err := service.GetAllUser()
	if err != nil {
		logger.Logger.WithField("actor", actor.Username).Errorf("authenticate/user/show: get all users: %v", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to retrieve users")
		return
	}
	if len(userList) == 0 {
		saveHistory(opHistory, "failure")
		response.NotFound(w, "no users found")
		return
	}

	var result []userShowAuthenticateResp
	for _, u := range userList {
		mappings, err := service.GetAllCliNeOfUserByUserId(u.AccountID)
		if err != nil {
			logger.Logger.WithField("user_id", u.AccountID).Errorf("authenticate/user/show: get ne mappings: %v", err)
		}
		var nes []tblNe
		for _, m := range mappings {
			ne, err := service.GetNeByNeId(m.TblNeID)
			if err != nil || ne == nil {
				continue
			}
			nes = append(nes, tblNe{Ne: ne.NeName, Site: ne.SiteName})
		}
		result = append(result, userShowAuthenticateResp{
			Username: u.AccountName,
			TblNes:   nes,
			Role:     service.GetPermissionByUser(u),
		})
	}

	saveHistory(opHistory, "success")
	response.Write(w, http.StatusFound, result)
}

// saveHistory is a fire-and-forget helper that logs DB errors without interrupting the caller.
func saveHistory(h db_models.CliOperationHistory, result string) {
	h.ExecutedTime = time.Now()
	h.Result = result
	if err := service.SaveHistoryCommand(h); err != nil {
		logger.Logger.Errorf("save history: %v", err)
	}
}

// ── local types ───────────────────────────────────────────────────────────────

type authUser struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type tokenRequestResponse struct {
	Status       string `json:"status"`
	ResponseData string `json:"response_data"`
	ResponseCode string `json:"response_code"`
	SystemType   string `json:"system_type"`
}

type userShowAuthenticateResp struct {
	Username string  `json:"username"`
	TblNes   []tblNe `json:"tblNes"`
	Role     string  `json:"role"`
}

type tblNe struct {
	Ne   string `json:"ne"`
	Site string `json:"site"`
}
