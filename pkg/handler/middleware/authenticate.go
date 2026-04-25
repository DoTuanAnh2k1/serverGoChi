package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/token"
)

// User is the authenticated caller extracted from the JWT.
type User struct {
	Username string
	Role     string // "super_admin", "admin", "user"
}

type contextKey string

const UserContextKey = contextKey("user")

// RequireAdmin rejects requests from users with role "user". Must be placed
// after Authenticate so that UserContextKey is populated.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, ok := r.Context().Value(UserContextKey).(*User)
		if !ok || u == nil {
			response.Unauthorized(w)
			return
		}
		if u.Role != "admin" && u.Role != "super_admin" {
			response.Write(w, http.StatusForbidden, "insufficient permissions: admin or super_admin required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Logger.Warn("middleware: authenticate: missing Authorization header")
			response.Unauthorized(w)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Basic" {
			logger.Logger.Warnf("middleware: authenticate: malformed Authorization header: %q", parts[0])
			response.Unauthorized(w)
			return
		}

		username, role, err := token.ParseToken(parts[1])
		if err != nil {
			logger.Logger.Errorf("middleware: authenticate: invalid token: %v", err)
			response.Unauthorized(w)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, &User{Username: username, Role: role})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
