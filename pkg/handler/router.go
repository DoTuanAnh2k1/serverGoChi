package handler

import (
	"fmt"
	"net/http"
	"os"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/config"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/middleware"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/store"

	"github.com/go-chi/chi"
)

var (
	RouterBasePath string
	Router         *chi.Mux
)

// Init wires up the v2 HTTP surface. The shape mirrors docs/api-v2.md and the
// frontend tabs: users, nes, commands, ne-access-groups, cmd-exec-groups,
// password-policy, access-list, history, authorize, config-backup.
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
	Router.Get("/metrics", HandlerMetrics)
	Router.Get("/admin", handlerFrontend)
	Router.Get("/docs", handlerSwaggerUI)
	Router.Get("/docs/openapi.yaml", handlerOpenAPISpec)

	Router.Route("/aa", func(r chi.Router) {
		r.Get("/health-check-db", HealthCheck)
		r.Post("/validate-token", HandlerValidateToken)

		// Rate-limited public login.
		r.With(middleware.RateLimit(middleware.LoginRateLimiter)).Post("/authenticate", HandlerAuthenticate)

		// History save is open (proxy/cli-gate pushes audit events without JWT).
		r.Post("/history/save", HandlerSaveHistory)

		// All remaining routes need a valid JWT.
		r.Group(func(p chi.Router) {
			p.Use(middleware.Authenticate)

			// Any authenticated user can do these.
			p.Post("/change-password", HandlerChangePassword)
			p.Get("/users", HandlerListUsers)
			p.Get("/users/{id}", HandlerGetUser)
			p.Get("/nes", HandlerListNEs)
			p.Get("/nes/{id}", HandlerGetNE)
			p.Get("/commands", HandlerListCommands)
			p.Get("/ne-access-groups", HandlerListNeAccessGroups)
			p.Get("/ne-access-groups/{id}/users", HandlerNeAccessGroupUsers)
			p.Get("/ne-access-groups/{id}/nes", HandlerNeAccessGroupNEs)
			p.Get("/cmd-exec-groups", HandlerListCmdExecGroups)
			p.Get("/cmd-exec-groups/{id}/users", HandlerCmdExecGroupUsers)
			p.Get("/cmd-exec-groups/{id}/commands", HandlerCmdExecGroupCommands)
			p.Get("/password-policy", HandlerGetPasswordPolicy)
			p.Get("/access-list", HandlerListAccessList)
			p.Post("/authorize/check", HandlerAuthorizeCheck)
			p.Get("/history", HandlerListHistory)
			p.Get("/config-backup/list", HandlerConfigBackupList)
			p.Get("/config-backup/{id}", HandlerConfigBackupGet)

			// Admin or super_admin required for write operations.
			p.Group(func(a chi.Router) {
				a.Use(middleware.RequireAdmin)

				a.Post("/users", HandlerCreateUser)
				a.Put("/users/{id}", HandlerUpdateUser)
				a.Delete("/users/{id}", HandlerDeleteUser)
				a.Post("/users/{id}/reset-password", HandlerAdminResetPassword)

				a.Post("/nes", HandlerCreateNE)
				a.Put("/nes/{id}", HandlerUpdateNE)
				a.Delete("/nes/{id}", HandlerDeleteNE)

				a.Post("/commands", HandlerCreateCommand)
				a.Put("/commands/{id}", HandlerUpdateCommand)
				a.Delete("/commands/{id}", HandlerDeleteCommand)

				a.Post("/ne-access-groups", HandlerCreateNeAccessGroup)
				a.Put("/ne-access-groups/{id}", HandlerUpdateNeAccessGroup)
				a.Delete("/ne-access-groups/{id}", HandlerDeleteNeAccessGroup)
				a.Post("/ne-access-groups/{id}/users", HandlerNeAccessGroupAddUser)
				a.Delete("/ne-access-groups/{id}/users/{user_id}", HandlerNeAccessGroupRemoveUser)
				a.Post("/ne-access-groups/{id}/nes", HandlerNeAccessGroupAddNE)
				a.Delete("/ne-access-groups/{id}/nes/{ne_id}", HandlerNeAccessGroupRemoveNE)

				a.Post("/cmd-exec-groups", HandlerCreateCmdExecGroup)
				a.Put("/cmd-exec-groups/{id}", HandlerUpdateCmdExecGroup)
				a.Delete("/cmd-exec-groups/{id}", HandlerDeleteCmdExecGroup)
				a.Post("/cmd-exec-groups/{id}/users", HandlerCmdExecGroupAddUser)
				a.Delete("/cmd-exec-groups/{id}/users/{user_id}", HandlerCmdExecGroupRemoveUser)
				a.Post("/cmd-exec-groups/{id}/commands", HandlerCmdExecGroupAddCommand)
				a.Delete("/cmd-exec-groups/{id}/commands/{command_id}", HandlerCmdExecGroupRemoveCommand)

				a.Put("/password-policy", HandlerUpsertPasswordPolicy)

				a.Post("/access-list", HandlerCreateAccessList)
				a.Delete("/access-list/{id}", HandlerDeleteAccessList)

				a.Post("/config-backup/save", HandlerConfigBackupSave)
			})
		})
	})
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := store.GetSingleton().Ping(); err != nil {
		response.InternalError(w, err.Error())
		return
	}
	response.Success(w, "Database still alive")
}

func handlerOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	specPath := os.Getenv("API_SPEC_PATH")
	if specPath == "" {
		specPath = "api.yaml"
	}
	data, err := os.ReadFile(specPath)
	if err != nil {
		response.NotFound(w, "api spec not found")
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func SwaggerUIHTML(specURL string) string {
	if specURL == "" {
		specURL = "/docs/openapi.yaml"
	}
	return `<!DOCTYPE html>
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
    url: "` + specURL + `",
    dom_id: '#swagger-ui',
    presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
    layout: "BaseLayout"
  });
</script>
</body>
</html>`
}

func handlerSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, SwaggerUIHTML("/docs/openapi.yaml"))
}
