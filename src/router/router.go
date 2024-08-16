package router

import (
	"net/http"
	"serverGoChi/src/router/authenticate"
	"serverGoChi/src/router/authorize"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/server"
	"serverGoChi/src/store"

	"github.com/go-chi/chi"
)

// RouterBasePath Variable
var RouterBasePath string

// Router Variable
var Router *chi.Mux

// Initialize Function in Router
func init() {
	// Initialize Router
	Router = chi.NewRouter()
	RouterBasePath = server.Config.GetString("ROUTER_BASE_PATH")

	// Set Router CORS Configuration
	middleware.RouterCORSCfg.Origins = server.Config.GetString("CORS_ALLOWED_ORIGIN")
	middleware.RouterCORSCfg.Methods = server.Config.GetString("CORS_ALLOWED_METHOD")
	middleware.RouterCORSCfg.Headers = server.Config.GetString("CORS_ALLOWED_HEADER")

	// Set Router Middleware
	Router.Use(middleware.RouterCORS)
	Router.Use(middleware.RouterRealIP)
	Router.Use(middleware.RouterEntitySize)

	// Set Router Handler
	Router.NotFound(handlerNotFound)
	Router.MethodNotAllowed(handlerMethodNotAllowed)
	Router.Get("/favicon.ico", handlerFavIcon)

	// Router.Post("/aa/authorize", handlerAuthorize)
	Router.Post("/aa/authenticate", authenticate.HandlerAuthenticate)
	Router.Post("/aa/authenticate/user/set", middleware.Authenticate(authenticate.HandlerAuthenticateUserSet))
	Router.Post("/aa/authenticate/user/delete", middleware.Authenticate(authenticate.HandlerAuthenticateUserDelete))
	Router.Get("/aa/authenticate/user/show", middleware.Authenticate(authenticate.HandlerAuthenticateUserShow))

	Router.Post("/aa/authorize/permission/set", middleware.Authenticate(authorize.HandlerPermissionSet))
	Router.Post("/aa/authorize/permission/delete", middleware.Authenticate(authorize.HandlerPermissionDelete))
	Router.Post("/aa/authorize/permission/show", middleware.Authenticate(authorize.HandlerPermissionShow))
}

// HealthCheck Function
func HealthCheck(w http.ResponseWriter) {
	// Check Database Connections

	err := store.GetSingleton().Ping()

	if err != nil {
		response.InternalError(w, err.Error())
	} else {
		response.Success(w, "")
	}
}
