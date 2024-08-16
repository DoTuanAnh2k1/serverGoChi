package authorize

import (
	"encoding/json"
	"net/http"
	"serverGoChi/src/log"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authorize"
)

func handlerAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var authorizeReq AuthorizeReq
	err := json.NewDecoder(r.Body).Decode(&authorizeReq)
	if err != nil {
		log.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}

	user, err := authorize.GetUserByName(authorizeReq.Ne)
	if err != nil {
		log.Logger.Error("Cant get user: ", err)
		response.Write(w, http.StatusBadRequest, "Cant get user")
		return
	}

	response.Write(w, http.StatusOK, user)
}

type AuthorizeReq struct {
	Ne         string `json:"ne"`
	Site       string `json:"site"`
	SystemType string `json:"system_type"`
}
