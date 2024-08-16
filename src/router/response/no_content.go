package response

import "net/http"

// NoContent Function
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(204)
}
