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

		// Everything else needs a valid JWT. v2 has no role check — access
		// decisions live in the service layer via NE access / cmd exec
		// groups, invoked via /authorize/check.
		r.Group(func(p chi.Router) {
			p.Use(middleware.Authenticate)

			p.Post("/change-password", HandlerChangePassword)

			p.Route("/users", func(u chi.Router) {
				u.Get("/", HandlerListUsers)
				u.Post("/", HandlerCreateUser)
				u.Get("/{id}", HandlerGetUser)
				u.Put("/{id}", HandlerUpdateUser)
				u.Delete("/{id}", HandlerDeleteUser)
				u.Post("/{id}/reset-password", HandlerAdminResetPassword)
			})

			p.Route("/nes", func(n chi.Router) {
				n.Get("/", HandlerListNEs)
				n.Post("/", HandlerCreateNE)
				n.Get("/{id}", HandlerGetNE)
				n.Put("/{id}", HandlerUpdateNE)
				n.Delete("/{id}", HandlerDeleteNE)
			})

			p.Route("/commands", func(c chi.Router) {
				c.Get("/", HandlerListCommands)
				c.Post("/", HandlerCreateCommand)
				c.Put("/{id}", HandlerUpdateCommand)
				c.Delete("/{id}", HandlerDeleteCommand)
			})

			p.Route("/ne-access-groups", func(g chi.Router) {
				g.Get("/", HandlerListNeAccessGroups)
				g.Post("/", HandlerCreateNeAccessGroup)
				g.Put("/{id}", HandlerUpdateNeAccessGroup)
				g.Delete("/{id}", HandlerDeleteNeAccessGroup)
				g.Get("/{id}/users", HandlerNeAccessGroupUsers)
				g.Post("/{id}/users", HandlerNeAccessGroupAddUser)
				g.Delete("/{id}/users/{user_id}", HandlerNeAccessGroupRemoveUser)
				g.Get("/{id}/nes", HandlerNeAccessGroupNEs)
				g.Post("/{id}/nes", HandlerNeAccessGroupAddNE)
				g.Delete("/{id}/nes/{ne_id}", HandlerNeAccessGroupRemoveNE)
			})

			p.Route("/cmd-exec-groups", func(g chi.Router) {
				g.Get("/", HandlerListCmdExecGroups)
				g.Post("/", HandlerCreateCmdExecGroup)
				g.Put("/{id}", HandlerUpdateCmdExecGroup)
				g.Delete("/{id}", HandlerDeleteCmdExecGroup)
				g.Get("/{id}/users", HandlerCmdExecGroupUsers)
				g.Post("/{id}/users", HandlerCmdExecGroupAddUser)
				g.Delete("/{id}/users/{user_id}", HandlerCmdExecGroupRemoveUser)
				g.Get("/{id}/commands", HandlerCmdExecGroupCommands)
				g.Post("/{id}/commands", HandlerCmdExecGroupAddCommand)
				g.Delete("/{id}/commands/{command_id}", HandlerCmdExecGroupRemoveCommand)
			})

			p.Route("/password-policy", func(pp chi.Router) {
				pp.Get("/", HandlerGetPasswordPolicy)
				pp.Put("/", HandlerUpsertPasswordPolicy)
			})

			p.Route("/access-list", func(al chi.Router) {
				al.Get("/", HandlerListAccessList)
				al.Post("/", HandlerCreateAccessList)
				al.Delete("/{id}", HandlerDeleteAccessList)
			})

			p.Route("/authorize", func(a chi.Router) {
				a.Post("/check", HandlerAuthorizeCheck)
			})

			p.Get("/history", HandlerListHistory)

			p.Route("/config-backup", func(cb chi.Router) {
				cb.Post("/save", HandlerConfigBackupSave)
				cb.Get("/list", HandlerConfigBackupList)
				cb.Get("/{id}", HandlerConfigBackupGet)
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
