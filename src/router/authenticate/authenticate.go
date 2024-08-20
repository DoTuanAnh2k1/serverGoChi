package authenticate

import (
	"encoding/json"
	"net/http"
	"serverGoChi/src/log"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authenticate"
	"serverGoChi/src/utils/request"
	jsonWebToken "serverGoChi/src/utils/token"
	"serverGoChi/src/utils/token_request_response"
)

func HandlerAuthenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var requestUser request.User
	err := json.NewDecoder(r.Body).Decode(&requestUser)
	if err != nil {
		log.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}

	log.Logger.Info("Handler authenticate with user: ", requestUser.UserName)
	checkPass, err, userId := authenticate.Authenticate(requestUser.UserName, requestUser.Password)
	if err != nil || checkPass == false {
		log.Logger.Error("Login failure: ", err)
		response.Unauthorized(w)
		return
	}

	roles, err := authenticate.GetRolesById(userId)
	if err != nil {
		log.Logger.Error("Cannot get role from userId: ", err)
		response.InternalError(w, "Check log for details")
		return
	}
	log.Logger.Info("Get user role: ", roles)

	tokenString, err := jsonWebToken.CreateToken(requestUser.UserName, roles)
	if err != nil {
		log.Logger.Error("Cannot create token with err: ", err)
		response.InternalError(w, "Cannot create token")
		return
	}
	log.Logger.Info("Create token for user: ", tokenString)

	err = authenticate.UpdateLoginHistory(requestUser.UserName, r.Host)
	if err != nil {
		log.Logger.Error("Error Update history login: ", err)
		response.Write(w, http.StatusInternalServerError, "Error Update history login")
		return
	}

	tokenReqResp := token_request_response.TokenRequestResponse{
		Status:       "success",
		ResponseData: tokenString,
		ResponseCode: "200",
		SystemType:   "5GC",
	}

	log.Logger.Infof("Requese comming from user %v with Ip Address: %v", requestUser.UserName, r.Host)
	response.Write(w, http.StatusOK, tokenReqResp)
}
