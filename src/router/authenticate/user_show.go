package authenticate

import (
	"fmt"
	"net/http"
	"regexp"
	"serverGoChi/models/db_models"
	"serverGoChi/src/log"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/service/authenticate"
	"serverGoChi/src/service/history_command"
	"serverGoChi/src/service/user"
	"time"
)

func HandlerAuthenticateUserShow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
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

	logOperationHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authenticate-user delete username %v password xxx", "xxx"), // Get from middleware
		CreatedDate: time.Now(),
		Scope:       "ext-config",
		Account:     userMiddleware.Username,
	}

	userList, err := user.GetAllUser()
	if err != nil {
		log.Logger.Error("Cant get all user from db: ", err)
		response.Write(w, http.StatusInternalServerError, "Cant get all user from db")
		return
	}
	if userList == nil || len(userList) == 0 {
		logOperationHistory.ExecutedTime = time.Now()
		logOperationHistory.Result = "failure"
		err = history_command.SaveHistoryCommand(logOperationHistory)
		if err != nil {
			log.Logger.Error("Cant save command to db: ", err)
		}

		response.NotFound(w, "Empty List of Users")
	}

	timePattern := "^[0-9]*$"
	pattern, err := regexp.Compile(timePattern)
	if err != nil {
		log.Logger.Error("Cant compile regex: ", err)
		response.Write(w, http.StatusInternalServerError, "Cant compile regex")
		return
	}

	var userShowAuthenticateRespList []UserShowAuthenticateResp
	for _, userElement := range userList {
		isMatches := pattern.MatchString(userElement.AccountName)
		if !isMatches || userElement.IsEnable {
			continue
		}

		tblId, err := authenticate.GetTblIdByUserId(userElement.AccountID)
		if err != nil {
			log.Logger.Error("Cant get tblId by user id: ", err)
			// response.Write(w, http.StatusInternalServerError, "Cant get tblId by user id")

			continue
		}

		neList, err := authenticate.GetNeListById(tblId)
		if err != nil {
			log.Logger.Error("Cant get ne list: ", err)
			// response.Write(w, http.StatusInternalServerError, "Cant get ne list")
			continue
		}

		var tblNes []TblNe
		if len(neList) != 0 {
			for _, ne := range neList {
				tblNes = append(tblNes, TblNe{
					Ne:   ne.Name,
					Site: ne.SiteName,
				})
			}
		}

		role, err := authenticate.GetRolesById(userElement.AccountID)
		if err != nil {
			log.Logger.Error("Cant get role by user id: ", err)
			// response.Write(w, http.StatusInternalServerError, "Cant get role by user id")
			continue
		}

		userShowAuthenticateRespList = append(userShowAuthenticateRespList, UserShowAuthenticateResp{
			Username: userElement.AccountName,
			TblNes:   tblNes,
			Role:     role,
		})
	}

	logOperationHistory.ExecutedTime = time.Now()
	logOperationHistory.Result = "success"
	err = history_command.SaveHistoryCommand(logOperationHistory)
	if err != nil {
		log.Logger.Error("Cant save command to db: ", err)
	}

	response.Write(w, http.StatusFound, userShowAuthenticateRespList)
}

type UserShowAuthenticateResp struct {
	Username string  `json:"username"`
	TblNes   []TblNe `json:"tblNes"`
	Role     string  `json:"role"`
}

type TblNe struct {
	Ne   string `json:"ne"`
	Site string `json:"site"`
}
