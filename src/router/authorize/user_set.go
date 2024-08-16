package authorize

import (
	"encoding/json"
	"net/http"
	"serverGoChi/src/log"
	"serverGoChi/src/router/response"
)

func HandlerUserSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var userSetReq UserSetReq
	err := json.NewDecoder(r.Body).Decode(&userSetReq)
	if err != nil {
		log.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}

	//logOperationHistory := db_models.CliOperationHistory{
	//	CmdName:     fmt.Sprintf("authorize-user set username %v permission %v", userSetReq.Username, userSetReq.Permission),
	//	CreatedDate: time.Now(),
	//	Scope:       "ext-config",
	//	Account:     "", // Get from middleware
	//}

}

type UserSetReq struct {
	Username   string `json:"username"`
	Permission string `json:"permission"`
}
