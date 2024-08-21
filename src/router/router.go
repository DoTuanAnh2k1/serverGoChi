package router

import (
	"net/http"
	"serverGoChi/config"
	"serverGoChi/src/router/authenticate"
	"serverGoChi/src/router/authorize"
	"serverGoChi/src/router/list"
	"serverGoChi/src/router/middleware"
	"serverGoChi/src/router/response"
	"serverGoChi/src/store"

	"github.com/go-chi/chi"
)

// RouterBasePath Variable
var RouterBasePath string

// Router Variable
var Router *chi.Mux

// Initialize Function in Router
func Init() {
	// Initialize Router
	routerCfg := config.GetRouterConfig()
	Router = chi.NewRouter()

	RouterBasePath = routerCfg.BasePath

	// Set Router CORS Configuration
	middleware.RouterCORSCfg.Origins = routerCfg.Origins
	middleware.RouterCORSCfg.Methods = routerCfg.Methods
	middleware.RouterCORSCfg.Headers = routerCfg.Headers

	// Set Router Middleware
	Router.Use(middleware.RouterCORS)
	Router.Use(middleware.RouterRealIP)

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
			r.Get("/ne/monitor", list.HandlerListNeMonitor)
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
