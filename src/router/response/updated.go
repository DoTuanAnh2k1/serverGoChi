package response

import "net/http"

// Updated Function
func Updated(w http.ResponseWriter) {
	var response ResSuccess

	// Set Response Data
	response.Status = true
	response.Code = http.StatusOK
	response.Message = "Updated"

	// Set Response Data to HTTP
	Write(w, response.Code, response)
}
