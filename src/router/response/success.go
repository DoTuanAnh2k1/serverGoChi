package response

import "net/http"

// Success Function
func Success(w http.ResponseWriter, message string) {
	var response ResSuccess

	// Set Default Message
	if len(message) == 0 {
		message = "Success"
	}

	// Set Response Data
	response.Status = true
	response.Code = http.StatusOK
	response.Message = message

	// Set Response Data to HTTP
	Write(w, response.Code, response)
}
