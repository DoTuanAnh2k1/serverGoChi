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

func HandlerListNEs(w http.ResponseWriter, r *http.Request) {
	nes, err := service.ListNEs()
	if err != nil {
		response.InternalError(w, err.Error())
		return
	}
	if nes == nil {
		nes = []*db_models.NE{}
	}
	response.Write(w, http.StatusOK, nes)
}

func HandlerCreateNE(w http.ResponseWriter, r *http.Request) {
	var n db_models.NE
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := service.CreateNE(&n); err != nil {
		response.Write(w, http.StatusBadRequest, err.Error())
		return
	}
	response.Write(w, http.StatusCreated, n)
}

func HandlerGetNE(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	n, err := service.GetNE(id)
	if err != nil {
		response.NotFound(w, err.Error())
		return
	}
	response.Write(w, http.StatusOK, n)
}

func HandlerUpdateNE(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	var n db_models.NE
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		response.Write(w, http.StatusBadRequest, "invalid request body")
		return
	}
	n.ID = id
	if err := service.UpdateNE(&n); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "updated")
}

func HandlerDeleteNE(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.Write(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := service.DeleteNE(id); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "deleted")
}
