package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/handler/response"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
	"github.com/DoTuanAnh2k1/serverGoChi/pkg/token"
)

// User for saving user information
type User struct {
	Username string
	Roles    string
}

// Context key type
type contextKey string

const UserContextKey = contextKey("user")

// Authenticate middleware
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

		userName, roles, err := token.ParseToken(parts[1])
		if err != nil {
			logger.Logger.Errorf("middleware: authenticate: invalid token: %v", err)
			response.Unauthorized(w)
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, &User{Username: userName, Roles: roles})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
