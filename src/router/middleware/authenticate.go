package middleware

import (
	"context"
	"encoding/base64"
	"net/http"
	"serverGoChi/src/log"
	"serverGoChi/src/router/response"
	"strings"
)

// User for saving user information
type User struct {
	Username string
}

// Context key type
type contextKey string

const UserContextKey = contextKey("user")

// Authenticate to check Basic Auth and save user context
func Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Logger.Error("Request don't have authorize header")
			response.Unauthorized(w)
			return
		}

		authHeaderParts := strings.SplitN(authHeader, " ", 2)
		if len(authHeaderParts) != 2 || authHeaderParts[0] != "Basic" {
			log.Logger.Error("Authorize header of request invalid")
			response.Unauthorized(w)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(authHeaderParts[1])
		if err != nil {
			log.Logger.Error("Cant decode authorize header of request")
			response.Unauthorized(w)
			return
		}

		parts := strings.SplitN(string(payload), ":", 2)
		if len(parts) != 2 || !validateCredentials(parts[0], parts[1]) {
			log.Logger.Info("Authorize header of request Fail")
			response.Unauthorized(w)
			return
		}

		user := &User{Username: parts[0]}
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next(w, r.WithContext(ctx))
	}
}

func validateCredentials(username, password string) bool {
	// Hard Code here, need to change
	// Tự dưng quên mẹ mất :D
	// ok chạy lại hộ phát
	// return username == "namnd27" && password == "1"
	return true
}
