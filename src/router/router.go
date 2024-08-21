package router

import (
	"net/http"
	"serverGoChi/src/router/authenticate"
	"serverGoChi/src/router/authorize"
	"serverGoChi/src/router/list"
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

	Router.Route("/aa", func(router chi.Router) {
		router.Post("/authenticate", authenticate.HandlerAuthenticate)
		router.Route("/authenticate/user", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Use(middleware.CheckRole)

			r.Post("/set", authenticate.HandlerAuthenticateUserSet)
			r.Post("/delete", authenticate.HandlerAuthenticateUserDelete)
			r.Get("/show", authenticate.HandlerAuthenticateUserShow)
		})

		router.Route("/authorize", func(r chi.Router) {
			r.Route("/permission", func(subRouter chi.Router) {
				subRouter.Use(middleware.Authenticate)
				subRouter.Use(middleware.CheckRole)

				subRouter.Post("/set", authorize.HandlerPermissionSet)
				subRouter.Post("/delete", authorize.HandlerPermissionDelete)
				subRouter.Get("/show", authorize.HandlerPermissionShow)
			})

			r.Route("/ne", func(subRouter chi.Router) {
				subRouter.Use(middleware.Authenticate)
				subRouter.Use(middleware.CheckRole)

				subRouter.Post("/delete", authorize.HandlerNeDelete)
				subRouter.Post("/set", authorize.HandlerNeSet)
				subRouter.Get("/show", authorize.HandlerNeShow)
			})

			r.Route("/user", func(subRouter chi.Router) {
				subRouter.Use(middleware.Authenticate)
				subRouter.Use(middleware.CheckRole)

				subRouter.Post("/set", authorize.HandlerUserSet)
				subRouter.Post("/delete", authorize.HandlerUserDelete)
				subRouter.Get("/show", authorize.HandlerUserShow)
			})
		})

		router.Route("/list", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Use(middleware.CheckRole)

			r.Get("/ne", list.HandlerListNe)
		})
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
