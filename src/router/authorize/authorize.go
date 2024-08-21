package authorize

import (
	"encoding/json"
	"net/http"
	"serverGoChi/src/log"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/user"
)

func HandlerAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userMiddleware, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		log.Logger.Error("Error to get user from token key")
		response.InternalError(w, "Internal Server Error")
		return
	}

	var authorizeReq AuthorizeReq
	err := json.NewDecoder(r.Body).Decode(&authorizeReq)
	if err != nil {
		log.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}
	log.Logger.Infof("Handler Authorize with user: %v", userMiddleware.Username)

	u, err := user.GetUserByUserName(userMiddleware.Username)
	if err != nil {
		log.Logger.Error("Cant get user: ", err)
		response.Write(w, http.StatusBadRequest, "Cant get user")
		return
	}

	response.Write(w, http.StatusOK, u)
}
