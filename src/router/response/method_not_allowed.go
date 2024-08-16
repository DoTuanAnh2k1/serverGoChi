package response

import "net/http"

// MethodNotAllowed Function
func MethodNotAllowed(w http.ResponseWriter, message string) {
	var response ResError

	// Set Default Message
	if len(message) == 0 {
		message = "Method Not Allowed"
	}

	// Set Response Data
	response.Status = false
	response.Code = http.StatusMethodNotAllowed
	response.Message = "Method Not Allowed"
	response.Error = message

	// Set Response Data to HTTP
	Write(w, response.Code, response)
}
