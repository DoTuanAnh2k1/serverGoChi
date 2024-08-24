package validate

import (
	"encoding/json"
	"errors"
	"net/http"
	"serverGoChi/src/logger"
	"serverGoChi/src/router/response"
	"serverGoChi/src/utils/token"
	"strings"
)

func HandlerValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var tokenReq TokenReq
	err := json.NewDecoder(r.Body).Decode(&tokenReq)
	if err != nil {
		logger.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}

	logger.Logger.Info("Handler validate token")

	tokenReqParts := strings.SplitN(tokenReq.Token, " ", 2)
	if len(tokenReqParts) != 2 || tokenReqParts[0] != "Basic" {
		logger.Logger.Info("Token invalid")
		response.Unauthorized(w)
		return
	}

	userName, roles, err := token.ParseToken(tokenReq.Token)
	if err != nil {
		if errors.Is(err, errors.New("invalid token")) {
			logger.Logger.Info("Token invalid, cause: ", err)
			response.Unauthorized(w)
			return
		}

		logger.Logger.Error("Cannot verify token: ", err)
		response.InternalError(w, "Cannot verify token")
		return
	}

	tokenResp := &TokenResp{
		Username: userName,
		Roles:    roles,
	}
	logger.Logger.Info("Validate token success for username: ", userName)
	response.Write(w, http.StatusOK, tokenResp)
	return
}

type TokenReq struct {
	Token string `json:"token"`
}

type TokenResp struct {
	Username string `json:"username"`
	Roles    string `json:"roles"`
}
