package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/DoTuanAnh2k1/serverGoChi/internal/config"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/store"

	"github.com/go-chi/chi"
)

// RouterBasePath Variable
var RouterBasePath string

// Router Variable
var Router *chi.Mux

// Init initializes the HTTP router
func Init() {
	routerCfg := config.GetRouterConfig()
	Router = chi.NewRouter()

	RouterBasePath = routerCfg.BasePath

	middleware.RouterCORSCfg.Origins = routerCfg.Origins
	middleware.RouterCORSCfg.Methods = routerCfg.Methods
	middleware.RouterCORSCfg.Headers = routerCfg.Headers

	Router.Use(middleware.RouterCORS)
	Router.Use(middleware.RouterRealIP)

	Router.NotFound(handlerNotFound)
	Router.MethodNotAllowed(handlerMethodNotAllowed)
	Router.Get("/favicon.ico", handlerFavIcon)

	Router.Get("/health", HealthCheck)
	Router.Get("/docs", handlerSwaggerUI)
	Router.Get("/docs/openapi.yaml", handlerOpenAPISpec)

	Router.Route("/aa", func(router chi.Router) {
		router.Get("/heath-check-db", HealthCheck)
		router.Post("/validate-token", HandlerValidateToken)

		router.Route("/change-password", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Use(middleware.CheckRole)
			r.Post("/", HandlerChangePassword)
		})

		router.With(middleware.RateLimit(middleware.LoginRateLimiter)).Post("/authenticate", HandlerAuthenticate)

		router.Route("/authenticate/user", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Use(middleware.CheckRole)

			r.Post("/set", HandlerAuthenticateUserSet)
			r.Post("/delete", HandlerAuthenticateUserDelete)
			r.Get("/show", HandlerAuthenticateUserShow)
		})

		router.Route("/authorize", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Use(middleware.CheckRole)

			r.Route("/permission", func(subRouter chi.Router) {
				subRouter.Post("/set", HandlerPermissionSet)
				subRouter.Post("/delete", HandlerPermissionDelete)
				subRouter.Get("/show", HandlerPermissionShow)
			})

			r.Route("/ne", func(subRouter chi.Router) {
				subRouter.Post("/delete", HandlerNeDelete)
				subRouter.Post("/set", HandlerNeSet)
				subRouter.Get("/show", HandlerNeShow)
			})

			r.Route("/user", func(subRouter chi.Router) {
				subRouter.Post("/set", HandlerAuthorizeUserSet)
				subRouter.Post("/delete", HandlerAuthorizeUserDelete)
				subRouter.Get("/show", HandlerAuthorizeUserShow)
			})
		})

		router.Route("/history", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Post("/save", HandlerSaveHistory)
		})

		router.Route("/list", func(r chi.Router) {
			r.Use(middleware.Authenticate)

			r.Get("/ne", HandlerListNe)
			r.Get("/ne/monitor", HandlerListNeMonitor)
		})

		router.Route("/subscribers", func(r chi.Router) {
			r.Use(middleware.Authenticate)

			r.Get("/files", HandlerListSubscriberFiles)
			r.Get("/files/{index}", HandlerViewSubscriberFile)
		})
	})
}

// HealthCheck Function
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	err := store.GetSingleton().Ping()
	if err != nil {
		response.InternalError(w, err.Error())
	} else {
		response.Success(w, "Database still alive")
	}
}

// handlerOpenAPISpec serves api.yaml
func handlerOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	specPath := os.Getenv("API_SPEC_PATH")
	if specPath == "" {
		specPath = "api.yaml"
	}
	data, err := os.ReadFile(specPath)
	if err != nil {
		response.InternalError(w, "api spec not found")
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// handlerSwaggerUI serves a Swagger UI page via CDN
func handlerSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, swaggerUIHTML)
}

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist/swagger-ui.css">
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist/swagger-ui-bundle.js"></script>
<script>
  SwaggerUIBundle({
    url: "/docs/openapi.yaml",
    dom_id: '#swagger-ui',
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
    layout: "BaseLayout"
  });
</script>
</body>
</html>`
