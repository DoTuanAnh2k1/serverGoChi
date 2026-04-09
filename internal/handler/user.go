package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-aa-server/internal/bcrypt"
	"go-aa-server/internal/handler/middleware"
	"go-aa-server/internal/handler/response"
	"go-aa-server/internal/logger"
	"go-aa-server/internal/service"
	"go-aa-server/models/db_models"
)

// HandlerChangePassword handles POST /aa/change-password/
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
		Scope:       "ext-config",
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

// HandlerAuthorizeUserSet handles POST /aa/authorize/user/set
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
		Scope:       "ext-config",
		Account:     actor.Username,
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

	roles, err := service.GetAllUserRolesMappingById(u.AccountID)
	if err != nil {
		log.Errorf("authorize/user/set: get user roles: %v", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to retrieve user roles")
		return
	}

	for _, role := range roles {
		if role.Permission == req.Permission {
			log.Warnf("authorize/user/set: permission %q already assigned", req.Permission)
			saveHistory(opHistory, "failure")
			response.Write(w, http.StatusNotModified, "permission already exists")
			return
		}
	}

	if err = service.AddUserRole(&db_models.CliRoleUserMapping{UserID: u.AccountID, Permission: req.Permission}); err != nil {
		log.Errorf("authorize/user/set: add role: %v", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to add permission")
		return
	}

	log.Info("authorize/user/set: permission added")
	saveHistory(opHistory, "success")
	response.Success(w, "permission added")
}

// HandlerAuthorizeUserDelete handles POST /aa/authorize/user/delete
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

	log := logger.Logger.WithField("actor", actor.Username).WithField("target", req.Username).WithField("permission", req.Permission)

	opHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authorize-user delete username %v permission %v", req.Username, req.Permission),
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     actor.Username,
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

	roles, err := service.GetAllUserRolesMappingById(u.AccountID)
	if err != nil {
		log.Errorf("authorize/user/delete: get user roles: %v", err)
		saveHistory(opHistory, "failure")
		response.InternalError(w, "failed to retrieve user roles")
		return
	}

	for _, role := range roles {
		if role.Permission != req.Permission {
			continue
		}
		if err = service.DeleteUserRole(&db_models.CliRoleUserMapping{UserID: u.AccountID, Permission: req.Permission}); err != nil {
			log.Errorf("authorize/user/delete: delete role: %v", err)
			saveHistory(opHistory, "failure")
			response.InternalError(w, "failed to delete permission")
			return
		}
		log.Info("authorize/user/delete: permission removed")
		saveHistory(opHistory, "success")
		response.Write(w, http.StatusOK, "permission removed")
		return
	}

	log.Warnf("authorize/user/delete: permission %q not found on user", req.Permission)
	saveHistory(opHistory, "failure")
	response.Write(w, http.StatusNotModified, "permission not found")
}

// HandlerAuthorizeUserShow handles GET /aa/authorize/user/show
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
		Scope:       "ext-config",
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
		roles, err := service.GetRolesById(a.AccountID)
		if err != nil {
			logger.Logger.WithField("user_id", a.AccountID).Errorf("authorize/user/show: get roles: %v", err)
		}
		result = append(result, userShowResp{Username: a.AccountName, Permissions: roles})
	}

	saveHistory(opHistory, "success")
	response.Write(w, http.StatusFound, result)
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
	Username   string `json:"username"`
	Permission string `json:"permission"`
}

type userShowResp struct {
	Username    string `json:"username"`
	Permissions string `json:"permissions"`
}
