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
	Router.Get("/metrics", HandlerMetrics)
	Router.Get("/admin", handlerFrontend)
	Router.Get("/docs", handlerSwaggerUI)
	Router.Get("/docs/openapi.yaml", handlerOpenAPISpec)

	Router.Route("/aa", func(router chi.Router) {
		router.Get("/heath-check-db", HealthCheck)
		router.Post("/validate-token", HandlerValidateToken)

		router.Route("/change-password", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Post("/", HandlerChangePassword)
		})

		router.With(middleware.RateLimit(middleware.LoginRateLimiter)).Post("/authenticate", HandlerAuthenticate)

		router.Route("/authenticate/user", func(r chi.Router) {
			r.Use(middleware.Authenticate)

			// Read — any authenticated user (Normal users need this to view directory)
			r.Get("/show", HandlerAuthenticateUserShow)

			// Write — admin only
			r.Group(func(g chi.Router) {
				g.Use(middleware.CheckRole)
				g.Post("/set", HandlerAuthenticateUserSet)
				g.Post("/delete", HandlerAuthenticateUserDelete)
				g.Post("/reset-password", HandlerAdminResetPassword)
			})
		})

		router.Route("/authorize", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Use(middleware.CheckRole)

			r.Route("/ne", func(subRouter chi.Router) {
				subRouter.Post("/create", HandlerNeCreate)
				subRouter.Post("/update", HandlerNeUpdate)
				subRouter.Post("/remove", HandlerNeRemove)
				subRouter.Post("/delete", HandlerNeDelete)
				subRouter.Post("/set", HandlerNeSet)
				subRouter.Get("/show", HandlerNeShow)

				subRouter.Route("/config", func(cfgRouter chi.Router) {
					cfgRouter.Post("/create", HandlerNeConfigCreate)
					cfgRouter.Get("/list", HandlerNeConfigList)
					cfgRouter.Post("/update", HandlerNeConfigUpdate)
					cfgRouter.Post("/delete", HandlerNeConfigDelete)
				})
			})

			r.Route("/user", func(subRouter chi.Router) {
				subRouter.Post("/set", HandlerAuthorizeUserSet)
				subRouter.Post("/delete", HandlerAuthorizeUserDelete)
				subRouter.Get("/show", HandlerAuthorizeUserShow)
			})
		})

		router.Route("/admin", func(r chi.Router) {
			r.Use(middleware.Authenticate)

			// Read — any authenticated user
			r.Get("/user/list", HandlerAdminUserList)
			r.Get("/user/full", HandlerAdminUserFullList)
			r.Get("/ne/list", HandlerAdminNeList)

			// Self-or-admin: handler enforces actor==target unless caller is admin.
			r.Post("/user/update", HandlerAdminUserUpdate)

			// Admin-only writes
			r.Group(func(g chi.Router) {
				g.Use(middleware.CheckRole)
				g.Post("/ne/create", HandlerAdminNeCreate)
				g.Post("/ne/update", HandlerAdminNeUpdate)
			})
		})

		router.Route("/group", func(r chi.Router) {
			r.Use(middleware.Authenticate)

			// Read — any authenticated user
			r.Get("/list", HandlerGroupList)
			r.Post("/show", HandlerGroupShow)
			r.Get("/user", HandlerUserGroupList)
			r.Get("/ne", HandlerGroupNeList)

			// Admin-only writes
			r.Group(func(g chi.Router) {
				g.Use(middleware.CheckRole)
				g.Post("/create", HandlerGroupCreate)
				g.Post("/update", HandlerGroupUpdate)
				g.Post("/delete", HandlerGroupDelete)
				g.Post("/user/assign", HandlerUserGroupAssign)
				g.Post("/user/unassign", HandlerUserGroupUnassign)
				g.Post("/ne/assign", HandlerGroupNeAssign)
				g.Post("/ne/unassign", HandlerGroupNeUnassign)
			})
		})

		router.Route("/import", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Use(middleware.CheckRole)
			r.Post("/", HandlerImport)
		})

		router.Route("/history", func(r chi.Router) {
			// /save is open: the SSH proxy and downstream services log commands
			// before the user's JWT is available (and forwarding the token adds
			// friction for a simple audit sink). The caller supplies the account
			// in the body. /list stays authenticated — reading the audit log is
			// an admin-side concern.
			r.Post("/save", HandlerSaveHistory)
			r.Group(func(g chi.Router) {
				g.Use(middleware.Authenticate)
				g.Get("/list", HandlerListHistory)
			})
		})

		router.Route("/list", func(r chi.Router) {
			r.Use(middleware.Authenticate)

			r.Get("/ne", HandlerListNe)
			r.Get("/ne/monitor", HandlerListNeMonitor)
			r.Get("/ne/config", HandlerListNeConfig)
		})

		router.Route("/subscribers", func(r chi.Router) {
			r.Use(middleware.Authenticate)

			r.Get("/files", HandlerListSubscriberFiles)
			r.Get("/files/{index}", HandlerViewSubscriberFile)
		})

		router.Route("/config-backup", func(r chi.Router) {
			r.Use(middleware.Authenticate)

			r.Post("/save", HandlerConfigBackupSave)
			r.Get("/list", HandlerConfigBackupList)
			r.Get("/{id}", HandlerConfigBackupGet)
		})

		// ── RBAC (docs/rbac-design.md) ────────────────────────────────────
		//
		// NE profiles classify NEs by command set (SMF/AMF/UPF/...).
		// Command defs are the registry of allowed patterns per profile.
		// Command groups bundle defs so permissions can reference whole
		// bundles. Group cmd-permissions tie a group to commands at a
		// specific ne_scope with an effect ("allow" | "deny"), evaluated
		// by the AWS-IAM + scope-specificity logic in service/rbac.
		router.Route("/ne-profile", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Get("/list", HandlerListNeProfiles)
			r.Group(func(g chi.Router) {
				g.Use(middleware.CheckRole)
				g.Post("/create", HandlerCreateNeProfile)
				g.Post("/update", HandlerUpdateNeProfile)
				g.Delete("/{id}", HandlerDeleteNeProfile)
			})
		})

		router.Route("/ne/{ne_id}/profile", func(r chi.Router) {
			r.Use(middleware.Authenticate, middleware.CheckRole)
			r.Post("/", HandlerAssignNeProfile)
		})

		router.Route("/command-def", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Get("/list", HandlerListCommandDefs)
			r.Group(func(g chi.Router) {
				g.Use(middleware.CheckRole)
				g.Post("/create", HandlerCreateCommandDef)
				g.Post("/update", HandlerUpdateCommandDef)
				g.Post("/import", HandlerImportCommandDefs)
				g.Delete("/{id}", HandlerDeleteCommandDef)
			})
		})

		router.Route("/command-group", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Get("/list", HandlerListCommandGroups)
			r.Get("/{id}/commands", HandlerListCommandsOfGroup)
			r.Group(func(g chi.Router) {
				g.Use(middleware.CheckRole)
				g.Post("/create", HandlerCreateCommandGroup)
				g.Post("/update", HandlerUpdateCommandGroup)
				g.Delete("/{id}", HandlerDeleteCommandGroup)
				g.Post("/{id}/commands", HandlerAddCommandToGroup)
				g.Delete("/{id}/commands/{cmd_id}", HandlerRemoveCommandFromGroup)
			})
		})

		router.Route("/group/{id}/cmd-permissions", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Get("/", HandlerListGroupCmdPermissions)
			r.Group(func(g chi.Router) {
				g.Use(middleware.CheckRole)
				g.Post("/", HandlerCreateGroupCmdPermission)
				g.Delete("/{perm_id}", HandlerDeleteGroupCmdPermission)
			})
		})

		// Authorize — queried by downstream ne-config / ne-command services
		// once per session (effective) or per-command (check-command). Both
		// need only a valid JWT; evaluation handles the allow/deny.
		router.Route("/authorize/rbac", func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Get("/effective", HandlerAuthorizeEffective)
			r.Post("/check-command", HandlerAuthorizeCheckCommand)
		})
	})
}

// HealthCheck kiểm tra kết nối đến database.
//
// Input : GET (không có body/query params)
// Output: 200 "Database still alive" nếu DB phản hồi
//         500 kèm error message nếu Ping thất bại
// Flow  : store.GetSingleton().Ping() → trả kết quả
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
		response.NotFound(w, "api spec not found")
		return
	}
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// SwaggerUIHTML returns the Swagger UI HTML page pointing to the given spec URL.
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

// handlerSwaggerUI serves a Swagger UI page via CDN
func handlerSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, SwaggerUIHTML("/docs/openapi.yaml"))
}

