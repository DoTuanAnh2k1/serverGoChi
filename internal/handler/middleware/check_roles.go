package middleware

import (
	"net/http"
	"strings"

	"github.com/DoTuanAnh2k1/serverGoChi/internal/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/internal/logger"
)

// CheckRole middleware — requires the "admin" role.
func CheckRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := r.Context().Value(UserContextKey).(*User)
		if !ok {
			logger.Logger.Error("middleware: check_role: user not found in context")
			response.InternalError(w, "Internal Server Error")
			return
		}

		for _, role := range strings.Split(u.Roles, " ") {
			if strings.EqualFold(role, "admin") {
				next.ServeHTTP(w, r)
				return
			}
		}

		logger.Logger.WithField("user", u.Username).Warnf("middleware: check_role: forbidden — roles=%q", u.Roles)
		response.Write(w, http.StatusForbidden, "No authority")
	})
}
