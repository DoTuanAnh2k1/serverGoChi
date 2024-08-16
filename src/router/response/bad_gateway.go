package response

import (
	"net/http"
	"serverGoChi/src/log"
	"strings"
)

// BadGateway Function
func BadGateway(w http.ResponseWriter, message string) {
	var response ResError

	// Set Default Message
	if len(message) == 0 {
		message = "Bad Gateway"
	}

	// Set Response Data
	response.Status = false
	response.Code = http.StatusBadGateway
	response.Message = "Bad Gateway"
	response.Error = message

	// Logging Error
	log.Println(log.LogLevelError, "http-access", strings.ToLower(message))

	// Set Response Data to HTTP
	Write(w, response.Code, response)
}
