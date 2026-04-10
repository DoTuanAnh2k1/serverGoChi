package handler

import (
	"net/http"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
)

func handlerNotFound(w http.ResponseWriter, r *http.Request) {
	logger.Println(logger.LogLevelWarn, "http-access", "not found method "+r.Method+" at URI "+r.RequestURI)
	response.NotFound(w, "not found method "+r.Method+" at URI "+r.RequestURI)
}

func handlerMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	logger.Println(logger.LogLevelWarn, "http-access", "not allowed method "+r.Method+" at URI "+r.RequestURI)
	response.MethodNotAllowed(w, "not allowed method "+r.Method+" at URI "+r.RequestURI)
}

func handlerFavIcon(w http.ResponseWriter, r *http.Request) {
	response.NoContent(w)
}
