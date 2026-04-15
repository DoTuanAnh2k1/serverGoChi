package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/bcrypt"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// HandlerChangePassword cho phép user tự thay đổi mật khẩu của mình.
//
// Input : POST body JSON { "username": string, "old_password": string, "new_password": string }
// Output: 200 "password changed" nếu thành công
//         400 nếu body không hợp lệ
//         401 nếu actor không phải chính user đó
//         403 nếu old_password sai
//         500 nếu lỗi DB khi lấy hoặc cập nhật user
func HandlerChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req userChangePasswordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Errorf("change-password: decode request body: %v", err)
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actor, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		logger.Logger.Error("change-password: user not found in context")
		response.InternalError(w, "Internal Server Error")
		return
	}

	log := logger.Logger.WithField("actor", actor.Username).WithField("target", req.Username)

	opHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("change-password for user %v", req.Username),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     actor.Username,
	}

	if actor.Username != req.Username {
		log.Warnf("change-password: permission denied — cannot change another user's password")
		saveHistory(opHistory, "failure")
		response.Unauthorized(w)
		return
	}

	u, err := service.GetUserByUserName(req.Username)
	if err != nil {
		log.Errorf("change-password: get user: %v", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to retrieve user")
		return
	}

	if !bcrypt.Matches(req.Username+req.OldPassword, u.Password) {
		log.Warn("change-password: old password mismatch")
		saveHistory(opHistory, "failure")
		response.Write(w, http.StatusForbidden, "wrong password")
		return
	}

	u.Password = bcrypt.Encode(req.Username + req.NewPassword)
	u.LockedTime = time.Now()
	if err = service.UpdateUser(u); err != nil {
		log.Errorf("change-password: update user: %v", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to update password")
		return
	}

	log.Info("change-password: success")
	saveHistory(opHistory, "success")
	response.Success(w, "password changed")
}

// HandlerAuthorizeUserSet sets the account_type (permission) for a user.
//
// Input : POST body JSON { "username": string, "permission": "admin"|"user" }
// Output: 200 "permission updated" nếu thành công
//         400 nếu permission không hợp lệ
//         404 nếu user không tồn tại
//         500 nếu lỗi DB
// Flow  : decode body → GetUserByUserName → set AccountType theo permission →
//         UpdateUser → ghi operation history
func HandlerAuthorizeUserSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req userSetReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Errorf("authorize/user/set: decode request body: %v", err)
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actor, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		logger.Logger.Error("authorize/user/set: user not found in context")
		response.InternalError(w, "Internal Server Error")
		return
	}

	log := logger.Logger.WithField("actor", actor.Username).WithField("target", req.Username).WithField("permission", req.Permission)

	opHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authorize-user set username %v permission %v", req.Username, req.Permission),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     actor.Username,
	}

	if req.Username == service.SeedUsername {
		saveHistory(opHistory, "failure")
		response.Write(w, http.StatusForbidden, "cannot modify system user")
		return
	}

	var newAccountType int32
	switch req.Permission {
	case "admin":
		newAccountType = 1
	case "user":
		newAccountType = 2
	default:
		log.Warnf("authorize/user/set: invalid permission %q", req.Permission)
		saveHistory(opHistory, "failure")
		response.Write(w, http.StatusBadRequest, "permission must be 'admin' or 'user'")
		return
	}

	u, err := service.GetUserByUserName(req.Username)
	if err != nil {
		log.Errorf("authorize/user/set: get user: %v", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to retrieve user")
		return
	}
	if u == nil {
		log.Warn("authorize/user/set: user not found")
		saveHistory(opHistory, "failure")
		response.NotFound(w, "user not found")
		return
	}

	u.AccountType = newAccountType
	u.UpdatedDate = time.Now()
	if err = service.UpdateUser(u); err != nil {
		log.Errorf("authorize/user/set: update user: %v", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to update permission")
		return
	}

	log.Info("authorize/user/set: permission updated")
	saveHistory(opHistory, "success")
	response.Success(w, "permission updated")
}

// HandlerAuthorizeUserDelete resets a user's permission to "user" (account_type=2).
//
// Input : POST body JSON { "username": string }
// Output: 200 "permission reset" nếu thành công
//         404 nếu user không tồn tại
//         500 nếu lỗi DB
func HandlerAuthorizeUserDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req userDeleteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Errorf("authorize/user/delete: decode request body: %v", err)
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}

	actor, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		logger.Logger.Error("authorize/user/delete: user not found in context")
		response.InternalError(w, "Internal Server Error")
		return
	}

	log := logger.Logger.WithField("actor", actor.Username).WithField("target", req.Username)

	opHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authorize-user delete username %v", req.Username),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     actor.Username,
	}

	if req.Username == service.SeedUsername {
		saveHistory(opHistory, "failure")
		response.Write(w, http.StatusForbidden, "cannot modify system user")
		return
	}

	u, err := service.GetUserByUserName(req.Username)
	if err != nil {
		log.Errorf("authorize/user/delete: get user: %v", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to retrieve user")
		return
	}
	if u == nil {
		log.Warn("authorize/user/delete: user not found")
		saveHistory(opHistory, "failure")
		response.NotFound(w, "user not found")
		return
	}

	u.AccountType = 2
	u.UpdatedDate = time.Now()
	if err = service.UpdateUser(u); err != nil {
		log.Errorf("authorize/user/delete: update user: %v", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to reset permission")
		return
	}

	log.Info("authorize/user/delete: permission reset to user")
	saveHistory(opHistory, "success")
	response.Success(w, "permission reset")
}

// HandlerAuthorizeUserShow lists all users and their current permission (derived from account_type).
//
// Input : GET (no body)
// Output: 302 [ { username, permission } ]
//         500 nếu lỗi DB
func HandlerAuthorizeUserShow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	actor, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		logger.Logger.Error("authorize/user/show: user not found in context")
		response.InternalError(w, "Internal Server Error")
		return
	}

	opHistory := db_models.CliOperationHistory{
		CmdName:     "authorize-user show",
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     actor.Username,
	}

	accounts, err := service.GetAllUser()
	if err != nil {
		logger.Logger.WithField("actor", actor.Username).Errorf("authorize/user/show: get all users: %v", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to retrieve users")
		return
	}

	var result []userShowResp
	for _, a := range accounts {
		result = append(result, userShowResp{
			Username:   a.AccountName,
			Permission: service.GetPermissionByUser(a),
		})
	}

	saveHistory(opHistory, "success")
	response.Write(w, http.StatusFound, result)
}

// HandlerAdminResetPassword allows admin to reset any user's password.
//
// Input : POST body JSON { "username": string, "new_password": string }
// Output: 200 "password reset" nếu thành công
//         400 nếu thiếu username/new_password
//         404 nếu user không tồn tại
//         500 nếu lỗi DB
func HandlerAdminResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username    string `json:"username"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" || req.NewPassword == "" {
		response.Write(w, http.StatusBadRequest, "username and new_password are required")
		return
	}

	actor, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		response.InternalError(w, "Internal Server Error")
		return
	}

	op := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("admin reset-password for user %v", req.Username),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     actor.Username,
	}

	u, err := service.GetUserByUserName(req.Username)
	if err != nil {
		logger.Logger.Errorf("admin reset-password: get user: %v", err)
		saveHistory(op, "failure")
		response.InternalError(w, "failed to retrieve user")
		return
	}
	if u == nil {
		saveHistory(op, "failure")
		response.NotFound(w, "user not found")
		return
	}

	u.Password = bcrypt.Encode(req.Username + req.NewPassword)
	u.LockedTime = time.Now()
	if err := service.UpdateUser(u); err != nil {
		logger.Logger.Errorf("admin reset-password: update user: %v", err)
		saveHistory(op, "failure")
		response.InternalError(w, "failed to reset password")
		return
	}

	saveHistory(op, "success")
	response.Success(w, "password reset")
}

// ── local types ───────────────────────────────────────────────────────────────

type userChangePasswordReq struct {
	Username    string `json:"username"`
	NewPassword string `json:"new_password"`
	OldPassword string `json:"old_password"`
}

type userSetReq struct {
	Username   string `json:"username"`
	Permission string `json:"permission"`
}

type userDeleteReq struct {
	Username string `json:"username"`
}

type userShowResp struct {
	Username   string `json:"username"`
	Permission string `json:"permission"`
}
