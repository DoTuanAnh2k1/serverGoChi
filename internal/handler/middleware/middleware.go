package middleware

import (
	"net/http"
	"strings"
)

// Router CORS Configuration Struct
type routerCORSConfig struct {
	Origins string
	Methods string
	Headers string
}

// Router CORS Configuration Variable
var RouterCORSCfg routerCORSConfig

// RouterCORS Function
func RouterCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", RouterCORSCfg.Origins)
		w.Header().Set("Access-Control-Allow-Methods", RouterCORSCfg.Methods)
		w.Header().Set("Access-Control-Allow-Headers", RouterCORSCfg.Headers)
		next.ServeHTTP(w, r)
	})
}

// RouterRealIP Function
func RouterRealIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if XForwardedFor := r.Header.Get(http.CanonicalHeaderKey("X-Forwarded-For")); XForwardedFor != "" {
			dataIndex := strings.Index(XForwardedFor, ", ")
			if dataIndex == -1 {
				dataIndex = len(XForwardedFor)
			}
			r.RemoteAddr = XForwardedFor[:dataIndex]
		} else if XRealIP := r.Header.Get(http.CanonicalHeaderKey("X-Real-IP")); XRealIP != "" {
			r.RemoteAddr = XRealIP
		}
		next.ServeHTTP(w, r)
	})
}
