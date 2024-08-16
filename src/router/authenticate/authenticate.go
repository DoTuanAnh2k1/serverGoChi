package authenticate

import (
	"encoding/json"
	"net/http"
	"serverGoChi/src/log"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authenticate"
	"serverGoChi/src/utils/o_auth_token"
	"serverGoChi/src/utils/request"
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

	token, err := o_auth_token.RequestAccessToken(requestUser)
	if err != nil {
		log.Logger.Error("Error Get Token: ", err)
		response.Write(w, http.StatusInternalServerError, "Error Get Token")
		return
	}

	err = authenticate.UpdateLoginHistory(requestUser.UserName, r.Host)
	if err != nil {
		log.Logger.Error("Error Update history login: ", err)
		response.Write(w, http.StatusInternalServerError, "Error Update history login")
		return
	}

	tokenReqResp := token_request_response.TokenRequestResponse{
		Status:       "success",
		ResponseData: token.AccessToken,
		ResponseCode: "200",
		SystemType:   "5GC",
	}
	log.Logger.Infof("Requese comming from user %v with Ip Address: %v", requestUser.UserName, r.Host)
	response.Write(w, http.StatusOK, tokenReqResp)
}
