package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
)

// HandlerPermissionSet handles POST /aa/authorize/permission/set
func HandlerPermissionSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var cliRole *db_models.CliRole
	err := json.NewDecoder(r.Body).Decode(&cliRole)
	if err != nil {
		logger.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}

	userMiddleware, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		logger.Logger.Error("Error to get user from token key")
		response.InternalError(w, "Internal Server Error")
		return
	}

	loggerOperationHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authorize-permission set permission %v scope %v ne %v include type %v path %v", cliRole.Permission, cliRole.Scope, cliRole.NeType, cliRole.IncludeType, cliRole.Path),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     userMiddleware.Username,
	}

	isExist, err := service.IsExistCliRole(cliRole)
	if err != nil {
		logger.Logger.Error("Cant check is exist cli role: ", err)
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err1 := service.SaveHistoryCommand(loggerOperationHistory)
		if err1 != nil {
			logger.Logger.Error("Cant save command to db: ", err1)
		}
		response.Write(w, http.StatusInternalServerError, "Cant check is exist cli role")
		return
	}
	if isExist {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err := service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		response.Write(w, http.StatusNotModified, "")
	} else {
		err = service.CreateCliRole(cliRole)
		if err != nil {
			loggerOperationHistory.ExecutedTime = time.Now()
			loggerOperationHistory.Result = "failure"
			err := service.SaveHistoryCommand(loggerOperationHistory)
			if err != nil {
				logger.Logger.Error("Cant save command to db: ", err)
			}
			logger.Logger.Error("Error create cli role: ", err)
			response.Write(w, http.StatusInternalServerError, "Error create cli role")
			return
		}
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "success"
		err := service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		response.Created(w)
	}
}

// HandlerPermissionDelete handles POST /aa/authorize/permission/delete
func HandlerPermissionDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logger.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var cliRole *db_models.CliRole
	err := json.NewDecoder(r.Body).Decode(&cliRole)
	if err != nil {
		logger.Logger.Error("Error parsing JSON request body: ", err)
		response.Write(w, http.StatusInternalServerError, "Error parsing JSON request body")
		return
	}

	userMiddleware, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		logger.Logger.Error("Error to get user from token key")
		response.InternalError(w, "Internal Server Error")
		return
	}

	loggerOperationHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authorize-permission delete permission %v scope %v ne %v include type %v path %v", cliRole.Permission, cliRole.Scope, cliRole.NeType, cliRole.IncludeType, cliRole.Path),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     userMiddleware.Username,
	}

	isExist, err := service.IsExistCliRole(cliRole)
	if err != nil {
		logger.Logger.Error("Cant check is exist cli role: ", err)
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err1 := service.SaveHistoryCommand(loggerOperationHistory)
		if err1 != nil {
			logger.Logger.Error("Cant save command to db: ", err1)
		}
		response.Write(w, http.StatusInternalServerError, "Cant check is exist cli role")
		return
	}
	if !isExist {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err := service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		response.Write(w, http.StatusNotModified, "")
	} else {
		err = service.DeleteCliRole(cliRole)
		if err != nil {
			loggerOperationHistory.ExecutedTime = time.Now()
			loggerOperationHistory.Result = "failure"
			err := service.SaveHistoryCommand(loggerOperationHistory)
			if err != nil {
				logger.Logger.Error("Cant save command to db: ", err)
			}
			logger.Logger.Error("Error Delete cli role: ", err)
			response.Write(w, http.StatusInternalServerError, "Error Delete cli role")
			return
		}
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "success"
		err := service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		response.Success(w, "")
	}
}

// HandlerPermissionShow handles GET /aa/authorize/permission/show
func HandlerPermissionShow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		logger.Logger.Error("Method not allowed")
		response.Write(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userMiddleware, ok := r.Context().Value(middleware.UserContextKey).(*middleware.User)
	if !ok {
		logger.Logger.Error("Error to get user from token key")
		response.InternalError(w, "Internal Server Error")
		return
	}

	loggerOperationHistory := db_models.CliOperationHistory{
		CmdName:     fmt.Sprintf("authorize-permission show"),
		CreatedDate: time.Now(),
		Scope:       "cli-config",
		Account:     userMiddleware.Username,
	}

	cliRoleList, err := service.GetAllCliRoles()
	if err != nil {
		logger.Logger.Error("Cant get cli role list: ", err)
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err := service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		response.Write(w, http.StatusInternalServerError, "Cant get cli role list")
		return
	}

	if cliRoleList == nil || len(cliRoleList) == 0 {
		loggerOperationHistory.ExecutedTime = time.Now()
		loggerOperationHistory.Result = "failure"
		err := service.SaveHistoryCommand(loggerOperationHistory)
		if err != nil {
			logger.Logger.Error("Cant save command to db: ", err)
		}
		response.Write(w, http.StatusNotFound, "cli role list empty")
		return
	}

	loggerOperationHistory.ExecutedTime = time.Now()
	loggerOperationHistory.Result = "success"
	err = service.SaveHistoryCommand(loggerOperationHistory)
	if err != nil {
		logger.Logger.Error("Cant save command to db: ", err)
	}
	response.Write(w, http.StatusFound, cliRoleList)
}
