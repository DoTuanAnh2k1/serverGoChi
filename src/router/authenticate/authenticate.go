package authenticate

import (
	"encoding/json"
	"net/http"
	"serverGoChi/src/logger"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authenticate"
	jsonWebToken "serverGoChi/src/utils/token"
)

func HandlerAuthenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var requestUser User
	err := json.NewDecoder(r.Body).Decode(&requestUser)
	if err != nil {
		logger.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}

	logger.Logger.Info("Handler authenticate with user: ", requestUser.UserName)
	checkPass, err, userId := authenticate.Authenticate(requestUser.UserName, requestUser.Password)
	if err != nil || checkPass == false {
		logger.Logger.Error("Login failure: ", err)
		response.Unauthorized(w)
		return
	}

	roles, err := authenticate.GetRolesById(userId)
	if err != nil {
		logger.Logger.Error("Cannot get role from userId: ", err)
		response.InternalError(w, "Check log for details")
		return
	}
	logger.Logger.Info("Get user role: ", roles)

	tokenString, err := jsonWebToken.CreateToken(requestUser.UserName, roles)
	if err != nil {
		logger.Logger.Error("Cannot create token with err: ", err)
		response.InternalError(w, "Cannot create token")
		return
	}
	logger.Logger.Info("Create token for user: ", tokenString)

	err = authenticate.UpdateLoginHistory(requestUser.UserName, r.Host)
	if err != nil {
		logger.Logger.Error("Error Update history login: ", err)
		response.Write(w, http.StatusInternalServerError, "Error Update history login")
		return
	}

	tokenReqResp := TokenRequestResponse{
		Status:       "success",
		ResponseData: tokenString,
		ResponseCode: "200",
		SystemType:   "5GC",
	}

	logger.Logger.Infof("Requese comming from user %v with Ip Address: %v", requestUser.UserName, r.Host)
	response.Write(w, http.StatusOK, tokenReqResp)
}

type TokenRequestResponse struct {
	Status       string `json:"status"`
	ResponseData string `json:"response_data"`
	ResponseCode string `json:"response_code"`
	SystemType   string `json:"system_type"`
}

// User stand for RequestUser
type User struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}
