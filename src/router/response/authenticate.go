package response

import (
	"net/http"
)

// Authenticate Function
func Authenticate(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Authorization Required"`)
	Unauthorized(w)
}
