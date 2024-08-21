package response

import (
	"net/http"
	"serverGoChi/src/logger"
	"strings"
)

// InternalError Function
func InternalError(w http.ResponseWriter, message string) {
	var response ResError

	// Set Default Message
	if len(message) == 0 {
		message = "Internal Server Error"
	}

	// Set Response Data
	response.Status = false
	response.Code = http.StatusInternalServerError
	response.Message = "Internal Server Error"
	response.Error = message

	// Logging Error
	logger.Println(logger.LogLevelError, "http-access", strings.ToLower(message))

	// Set Response Data to HTTP
	Write(w, response.Code, response)
}
