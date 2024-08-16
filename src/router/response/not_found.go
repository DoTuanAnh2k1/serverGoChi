package response

import "net/http"

// NotFound Function
func NotFound(w http.ResponseWriter, message string) {
	var response ResError

	// Set Default Message
	if len(message) == 0 {
		message = "Not Found"
	}

	// Set Response Data
	response.Status = false
	response.Code = http.StatusNotFound
	response.Message = "Not Found"
	response.Error = message

	// Set Response Data to HTTP
	Write(w, response.Code, response)
}
