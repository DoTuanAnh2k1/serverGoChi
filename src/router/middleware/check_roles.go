package middleware

import (
	"net/http"
	"serverGoChi/src/logger"
	"serverGoChi/src/router/response"
	"strings"
)

// CheckRole middleware
func CheckRole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user from the context
		u, ok := r.Context().Value(UserContextKey).(*User)
		if !ok {
			logger.Logger.Error("Failed to retrieve user from context")
			response.InternalError(w, "Internal Server Error")
			return
		}

		logger.Logger.Info("User retrieved from context: ", u)

		// Check if the user has an "admin" role
		checkRole := false
		roleList := strings.Split(u.Roles, " ")
		for _, role := range roleList {
			if strings.EqualFold(role, "admin") {
				checkRole = true
				break
			}
		}

		// If the user does not have the required role, return a forbidden status
		if !checkRole {
			logger.Logger.Error("User does not have the required role")
			response.Write(w, http.StatusForbidden, "No authority")
			return
		}

		// Proceed to the next handler if the role check passes
		next.ServeHTTP(w, r)
	})
}
