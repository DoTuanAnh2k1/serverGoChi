package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/go-chi/chi"
)

type createUserReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

type updateUserReq struct {
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Phone     string `json:"phone"`
	IsEnabled *bool  `json:"is_enabled,omitempty"`
	Role      string `json:"role,omitempty"`
}

// callerFromRequest extracts the authenticated user from the request context.
func callerFromRequest(r *http.Request) *middleware.User {
	u, _ := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	return u
}

// requireAdmin returns true if caller is admin or super_admin; writes 403 otherwise.
func requireAdmin(w http.ResponseWriter, r *http.Request) (*middleware.User, bool) {
	caller := callerFromRequest(r)
	if caller == nil {
		response.Unauthorized(w)
		return nil, false
	}
	if caller.Role != "admin" && caller.Role != "super_admin" {
		response.Write(w, http.StatusForbidden, "insufficient permissions: admin or super_admin required")
		return nil, false
	}
	return caller, true
}

func HandlerListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := service.ListUsers()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	if users == nil {
		users = []*db_models.User{}
	}
	response.Write(w, http.StatusOK, users)
}

func HandlerCreateUser(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var req createUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	role := req.Role
	if role == "" {
		role = db_models.RoleUser
	}
	if role == db_models.RoleSuperAdmin && caller.Role != db_models.RoleSuperAdmin {
		response.Write(w, http.StatusForbidden, "only super_admin can create super_admin accounts")
		return
	}
	u := &db_models.User{
		Username: req.Username,
		Email:    req.Email,
		FullName: req.FullName,
		Phone:    req.Phone,
		Role:     role,
	}
	if err := service.CreateUser(u, req.Password); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Write(w, http.StatusCreated, u)
}

func HandlerGetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	u, err := service.GetUser(id)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}
	response.Write(w, http.StatusOK, u)
}

func HandlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	target, err := service.GetUser(id)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}
	if target.Role == db_models.RoleSuperAdmin && caller.Role != db_models.RoleSuperAdmin {
		response.Write(w, http.StatusForbidden, "only super_admin can modify super_admin accounts")
		return
	}
	var req updateUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	newRole := req.Role
	if newRole == db_models.RoleSuperAdmin && caller.Role != db_models.RoleSuperAdmin {
		response.Write(w, http.StatusForbidden, "only super_admin can promote to super_admin")
		return
	}
	if err := service.UpdateUserProfile(id, req.Email, req.FullName, req.Phone, req.IsEnabled, newRole); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(w, "updated")
}

func HandlerDeleteUser(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	target, err := service.GetUser(id)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}
	if target.Role == db_models.RoleSuperAdmin && caller.Role != db_models.RoleSuperAdmin {
		response.Write(w, http.StatusForbidden, "only super_admin can delete super_admin accounts")
		return
	}
	if err := service.DeleteUser(id); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "deleted")
}

type adminResetPasswordReq struct {
	NewPassword string `json:"new_password"`
}

func HandlerAdminResetPassword(w http.ResponseWriter, r *http.Request) {
	caller, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	target, err := service.GetUser(id)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}
	if target.Role == db_models.RoleSuperAdmin && caller.Role != db_models.RoleSuperAdmin {
		response.Write(w, http.StatusForbidden, "only super_admin can reset super_admin passwords")
		return
	}
	var req adminResetPasswordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := service.AdminResetPassword(id, req.NewPassword); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(w, "password reset")
}
