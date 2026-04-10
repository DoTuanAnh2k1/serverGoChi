package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DoTuanAnh2k1/serverGoChi/internal/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/token"
)

// HandlerValidateToken handles POST /aa/validate-token
func HandlerValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req validateTokenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Logger.Errorf("validate-token: decode request body: %v", err)
		response.Write(w, http.StatusInternalServerError, "invalid request body")
		return
	}

	parts := strings.SplitN(req.Token, " ", 2)
	if len(parts) != 2 || parts[0] != "Basic" {
		logger.Logger.Warn("validate-token: missing or malformed Basic prefix")
		response.Unauthorized(w)
		return
	}

	userName, roles, err := token.ParseToken(req.Token)
	if err != nil {
		logger.Logger.Errorf("validate-token: parse token: %v", err)
		response.Unauthorized(w)
		return
	}

	logger.Logger.WithField("user", userName).Debug("validate-token: ok")
	response.Write(w, http.StatusOK, &validateTokenResp{Username: userName, Roles: roles})
}

type validateTokenReq struct {
	Token string `json:"token"`
}

type validateTokenResp struct {
	Username string `json:"username"`
	Roles    string `json:"roles"`
}
