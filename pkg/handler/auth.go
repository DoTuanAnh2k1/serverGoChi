package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/token"
)

type authenticateReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authenticateResp struct {
	Status string `json:"status"`
	Token  string `json:"token"`
}

func HandlerAuthenticate(w http.ResponseWriter, r *http.Request) {
	var req authenticateReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Username) == "" || req.Password == "" {
		response.Write(w, http.StatusBadRequest, "username and password required")
		return
	}
	tok, err := service.Authenticate(req.Username, req.Password, clientIP(r))
	if err != nil {
		logger.Logger.WithField("user", req.Username).Warnf("auth: %v", err)
		// Don't leak which failure — auth gates are intentionally opaque.
		switch {
		case errors.Is(err, service.ErrAccountLocked),
			errors.Is(err, service.ErrAccountDisabled),
			errors.Is(err, service.ErrAccessDenied),
			errors.Is(err, service.ErrPasswordExpired):
			response.Write(w, http.StatusForbidden, err.Error())
		default:
			response.Unauthorized(w)
		}
		return
	}
	response.Write(w, http.StatusOK, authenticateResp{Status: "ok", Token: tok})
}

type validateTokenReq struct {
	Token string `json:"token"`
}

func HandlerValidateToken(w http.ResponseWriter, r *http.Request) {
	var req validateTokenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	username, err := token.ParseToken(req.Token)
	if err != nil {
		response.Unauthorized(w)
		return
	}
	response.Write(w, http.StatusOK, map[string]string{"status": "ok", "username": username})
}

type changePasswordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func HandlerChangePassword(w http.ResponseWriter, r *http.Request) {
	u, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok || u == nil {
		response.Unauthorized(w)
		return
	}
	var req changePasswordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	target, err := service.GetUserByUsername(u.Username)
	if err != nil {
		response.NotFound(w, "user not found")
		return
	}
	if err := service.ChangePassword(target.ID, req.OldPassword, req.NewPassword); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(w, "password changed")
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		if i := strings.Index(ip, ","); i > 0 {
			return strings.TrimSpace(ip[:i])
		}
		return strings.TrimSpace(ip)
	}
	if i := strings.LastIndex(r.RemoteAddr, ":"); i > 0 {
		return r.RemoteAddr[:i]
	}
	return r.RemoteAddr
}
