package response

import (
	"net/http"
	"serverGoChi/src/logger"
	"strings"
)

// BadRequest Function
func BadRequest(w http.ResponseWriter, message string) {
	var response ResError

	// Set Default Message
	if len(message) == 0 {
		message = "Bad Request"
	}

	// Set Response Data
	response.Status = false
	response.Code = http.StatusBadRequest
	response.Message = "Bad Request"
	response.Error = message

	// Logging Error
	logger.Println(logger.LogLevelError, "http-access", strings.ToLower(message))

	// Set Response Data to HTTP
	Write(w, response.Code, response)
}
