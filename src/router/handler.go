package router

import (
	"net/http"
	"serverGoChi/src/logger"
	"serverGoChi/src/router/response"
)

// HandlerNotFound Function
func handlerNotFound(w http.ResponseWriter, r *http.Request) {
	logger.Println(logger.LogLevelWarn, "http-access", "not found method "+r.Method+" at URI "+r.RequestURI)
	response.NotFound(w, "not found method "+r.Method+" at URI "+r.RequestURI)
}

// HandlerMethodNotAllowed Function
func handlerMethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	logger.Println(logger.LogLevelWarn, "http-access", "not allowed method "+r.Method+" at URI "+r.RequestURI)
	response.MethodNotAllowed(w, "not allowed method "+r.Method+" at URI "+r.RequestURI)
}

// HandlerFavIcon Function
func handlerFavIcon(w http.ResponseWriter, r *http.Request) {
	response.NoContent(w)
}
