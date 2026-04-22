package handler

import (
	"encoding/json"
	"fmt"
	"io"
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
//         nếu đang bị disable thì merge non-empty fields vào bản ghi cũ, bật is_enable
//         và UpdateUser (admin required fields được validate trên kết quả merged) →
//         ghi operation history
func HandlerAuthenticateUserSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Decode body that carries an optional group_ids list alongside the TblAccount fields.
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	var userInfo db_models.TblAccount
	if err := json.Unmarshal(raw, &userInfo); err != nil {
		logger.Logger.Errorf("authenticate/user/set: decode request body: %v", err)
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	var extra struct {
		GroupIDs []int64 `json:"group_ids"`
	}
	_ = json.Unmarshal(raw, &extra)

	if strings.TrimSpace(userInfo.AccountName) == "" {
		response.Write(w, http.StatusBadRequest, "username is required")
		return
	}
	if strings.TrimSpace(userInfo.Password) == "" {
		response.Write(w, http.StatusBadRequest, "password is required")
		return
	}
	// Baseline format + uniqueness checks. EnsureEmailUnique skips disabled
	// accounts so a fresh username can reuse an email freed by a disabled
	// account, and the re-enable path doesn't conflict with its own record.
	if err := service.ValidateUserCommon(&userInfo, userInfo.AccountName); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	// Prevent elevation to SuperAdmin via this endpoint.
	if userInfo.AccountType == 0 {
		response.Write(w, http.StatusForbidden, "cannot create SuperAdmin account")
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

	if existing == nil {
		// NEW USER PATH
		// Admin required-fields check (full_name, phone, ≥1 group) applies only
		// when creating fresh — re-enable validates against the merged result
		// below so admins aren't forced to resend fields already on file.
		if err := service.ValidateAdminUserFields(&userInfo, extra.GroupIDs); err != nil {
			response.Write(w, http.StatusBadRequest, err.Error())
			return
		}

		userInfo.Password = hashPass
		userInfo.IsEnable = true
		userInfo.CreatedBy = actor.Username
		now := time.Now()
		userInfo.CreatedDate, userInfo.UpdatedDate = now, now
		userInfo.LastLoginTime, userInfo.LastChangePass, userInfo.LockedTime = now, now, now
		if userInfo.AccountType != 1 && userInfo.AccountType != 2 {
			userInfo.AccountType = 2
		}
		userInfo.Status = true

		if err := service.AddUser(&userInfo); err != nil {
			log.Errorf("authenticate/user/set: create user: %v", err)
			saveHistory(opHistory, "failure")
			response.InternalError(w, "failed to create user")
			return
		}
		if len(extra.GroupIDs) > 0 {
			created, _ := service.GetUserByUserName(userInfo.AccountName)
			if created != nil {
				for _, gid := range extra.GroupIDs {
					if err := service.AssignUserToGroup(created.AccountID, gid); err != nil {
						log.Errorf("authenticate/user/set: assign group %d: %v", gid, err)
					}
				}
			}
		}
		log.Info("authenticate/user/set: user created")
		saveHistory(opHistory, "success")
		response.Created(w)
		return
	}

	if !existing.IsEnable {
		// RE-ENABLE PATH: merge only non-empty fields from the request so the
		// caller can pass just the fields they want to update. Password is
		// always refreshed (CLI requires it in the request).
		existing.Password = hashPass
		if v := strings.TrimSpace(userInfo.Email); v != "" {
			existing.Email = v
		}
		if v := strings.TrimSpace(userInfo.FullName); v != "" {
			existing.FullName = v
		}
		if v := strings.TrimSpace(userInfo.PhoneNumber); v != "" {
			existing.PhoneNumber = v
		}
		if v := strings.TrimSpace(userInfo.Address); v != "" {
			existing.Address = v
		}
		if v := strings.TrimSpace(userInfo.Description); v != "" {
			existing.Description = v
		}
		if userInfo.AccountType == 1 || userInfo.AccountType == 2 {
			existing.AccountType = userInfo.AccountType
		}
		existing.IsEnable = true
		existing.Status = true
		existing.CreatedBy = actor.Username
		existing.UpdatedDate = time.Now()
		existing.LoginFailureCount = 0

		// Validate merged result for admin accounts — full_name + phone
		// required regardless of whether they came from the request or were
		// already on file. Group requirement is NOT re-enforced here: prior
		// group mappings persist; the caller can optionally add more.
		if service.IsAdminAccountType(existing.AccountType) {
			if strings.TrimSpace(existing.FullName) == "" {
				response.Write(w, http.StatusBadRequest, "full_name is required for admin users")
				return
			}
			if strings.TrimSpace(existing.PhoneNumber) == "" {
				response.Write(w, http.StatusBadRequest, "phone_number is required for admin users")
				return
			}
			if err := service.ValidatePhone(existing.PhoneNumber); err != nil {
				response.Write(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		if err := service.UpdateUser(existing); err != nil {
			log.Errorf("authenticate/user/set: re-enable user: %v", err)
			saveHistory(opHistory, "failure")
			response.InternalError(w, "failed to enable user")
			return
		}
		// Additive group assignments (re-enable doesn't clear prior mappings).
		if len(extra.GroupIDs) > 0 {
			for _, gid := range extra.GroupIDs {
				if err := service.AssignUserToGroup(existing.AccountID, gid); err != nil {
					log.Errorf("authenticate/user/set: assign group %d: %v", gid, err)
				}
			}
		}
		log.Info("authenticate/user/set: user re-enabled with merged fields")
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

	// Guard: SuperAdmin accounts (account_type=0) cannot be disabled by anyone.
	if u != nil && u.AccountType == 0 {
		log.Warn("authenticate/user/delete: attempt to disable SuperAdmin — rejected")
		saveHistory(opHistory, "failure")
		response.Write(w, http.StatusForbidden, "cannot disable SuperAdmin account")
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
	// Note: SuperAdmin is intentionally NOT filtered here. This endpoint drives
	// the cli-netconf bootstrap (entrypoint-netconf.sh) which creates Linux
	// accounts for every returned user — hiding SuperAdmin would break SSH
	// logins for the seed account. Frontend listings use /admin/user/list which
	// does filter.
	if len(userList) == 0 {
		saveHistory(opHistory, "failure")
		response.NotFound(w, "no users found")
		return
	}

	var result []userShowAuthenticateResp
	for _, u := range userList {
		directMappings, err := service.GetAllCliNeOfUserByUserId(u.AccountID)
		if err != nil {
			logger.Logger.WithField("user_id", u.AccountID).Errorf("authenticate/user/show: get ne mappings: %v", err)
		}
		directSet := make(map[int64]struct{}, len(directMappings))
		var nes []tblNe
		for _, m := range directMappings {
			directSet[m.TblNeID] = struct{}{}
			ne, err := service.GetNeByNeId(m.TblNeID)
			if err != nil || ne == nil {
				continue
			}
			nes = append(nes, tblNe{Ne: ne.NeName, Site: ne.SiteName, ID: ne.ID, Namespace: ne.Namespace, ViaGroup: false})
		}

		reachable, err := service.GetAllNeIdsOfUser(u.AccountID)
		if err != nil {
			logger.Logger.WithField("user_id", u.AccountID).Errorf("authenticate/user/show: get union ne ids: %v", err)
		}
		for _, neId := range reachable {
			if _, ok := directSet[neId]; ok {
				continue
			}
			ne, err := service.GetNeByNeId(neId)
			if err != nil || ne == nil {
				continue
			}
			nes = append(nes, tblNe{Ne: ne.NeName, Site: ne.SiteName, ID: ne.ID, Namespace: ne.Namespace, ViaGroup: true})
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
	Ne        string `json:"ne"`
	Site      string `json:"site"`
	ID        int64  `json:"id"`
	Namespace string `json:"namespace"`
	ViaGroup  bool   `json:"via_group"`
}
