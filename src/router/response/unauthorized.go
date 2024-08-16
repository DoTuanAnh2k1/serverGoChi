package response

import "net/http"

// Unauthorized Function
func Unauthorized(w http.ResponseWriter) {
	var response ResError

	// Set Response Data
	response.Status = false
	response.Code = http.StatusUnauthorized
	response.Message = "Unauthorized"
	response.Error = "Unauthorized"

	// Set Response Data to HTTP
	Write(w, response.Code, response)
}
