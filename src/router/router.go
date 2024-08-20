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

	Router.Route("/aa/authenticate/user", func(r chi.Router) {
		r.Use(middleware.Authenticate)
		r.Use(middleware.CheckRole)

		r.Post("/set", authenticate.HandlerAuthenticateUserSet)
		r.Post("/delete", authenticate.HandlerAuthenticateUserDelete)
		r.Get("/show", authenticate.HandlerAuthenticateUserShow)
	})

	Router.Route("/aa/authorize/permission", func(r chi.Router) {
		r.Use(middleware.Authenticate)
		r.Use(middleware.CheckRole)

		r.Post("/set", authorize.HandlerPermissionSet)
		r.Post("/delete", authorize.HandlerPermissionDelete)
		r.Get("/show", authorize.HandlerPermissionShow)
	})

	Router.Route("/aa/authorize/ne", func(r chi.Router) {
		r.Use(middleware.Authenticate)
		r.Use(middleware.CheckRole)

		r.Get("/show", authorize.HandlerNeShow)
	})
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
