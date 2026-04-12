package handler

import (
	_ "embed"
	"net/http"
	"strings"
)

//go:embed frontend.html
var frontendHTML string

func handlerFrontend(w http.ResponseWriter, r *http.Request) {
	// Embedded mode: API is on same origin at /aa
	html := strings.ReplaceAll(frontendHTML, "{{API_BASE}}", "/aa")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}
