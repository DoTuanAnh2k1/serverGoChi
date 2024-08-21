package middleware

import (
	"context"
	"net/http"
	"serverGoChi/src/logger"
	"serverGoChi/src/router/response"
	"serverGoChi/src/utils/token"
	"strings"
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
			logger.Logger.Error("Request don't have authorize header")
			response.Unauthorized(w)
			return
		}

		authHeaderParts := strings.SplitN(authHeader, " ", 2)
		if len(authHeaderParts) != 2 || authHeaderParts[0] != "Basic" {
			logger.Logger.Error("Authorize header of request invalid")
			response.Unauthorized(w)
			return
		}

		tokenString := authHeaderParts[1]
		userName, roles, err := token.ParseToken(tokenString)
		if err != nil {
			logger.Logger.Error("Cannot verify token: ", err)
			response.InternalError(w, "Cannot verify token")
			return
		}

		user := &User{
			Username: userName,
			Roles:    roles,
		}
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
