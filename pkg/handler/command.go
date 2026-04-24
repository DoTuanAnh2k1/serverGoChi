package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/DoTuanAnh2k1/serverGoChi/models/db_models"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/service"
	"github.com/go-chi/chi"
)

func HandlerListCommands(w http.ResponseWriter, r *http.Request) {
	var neID int64
	if s := r.URL.Query().Get("ne_id"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			neID = n
		}
	}
	service_ := r.URL.Query().Get("service")
	cmds, err := service.ListCommands(neID, service_)
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	if cmds == nil {
		cmds = []*db_models.Command{}
	}
	response.Write(w, http.StatusOK, cmds)
}

func HandlerCreateCommand(w http.ResponseWriter, r *http.Request) {
	var c db_models.Command
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := service.CreateCommand(&c); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Write(w, http.StatusCreated, c)
}

func HandlerUpdateCommand(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	var c db_models.Command
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	c.ID = id
	if err := service.UpdateCommand(&c); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "updated")
}

func HandlerDeleteCommand(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := service.DeleteCommand(id); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "deleted")
}
