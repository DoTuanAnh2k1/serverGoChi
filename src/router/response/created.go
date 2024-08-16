package response

import "net/http"

// Created Function
func Created(w http.ResponseWriter) {
	var response ResSuccess

	// Set Response Data
	response.Status = true
	response.Code = http.StatusCreated
	response.Message = "Created"

	// Set Response Data to HTTP
	Write(w, response.Code, response)
}
