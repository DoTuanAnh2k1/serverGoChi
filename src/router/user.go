package router

import (
	"net/http"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/user"
)

func handlerUser(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	u, err := user.GetUserByUserName(name)
	if u == nil || err != nil {
		response.Write(w, http.StatusInternalServerError, "User not found")
		return
	}

	response.Write(w, http.StatusOK, u)
}
