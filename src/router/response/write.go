package response

import (
	"encoding/json"
	"net/http"
)

// Write Function
func Write(w http.ResponseWriter, responseCode int, responseData interface{}) {
	// Write Response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)

	// Write JSON to Response
	json.NewEncoder(w).Encode(responseData)
}
