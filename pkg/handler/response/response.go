package response

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DoTuanAnh2k1/serverGoChi/pkg/logger"
)

// ResSuccess Struct
type ResSuccess struct {
	Status  bool   `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ResError Struct
type ResError struct {
	Status  bool   `json:"status"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   string `json:"error"`
}

// Write Function
func Write(w http.ResponseWriter, responseCode int, responseData interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	json.NewEncoder(w).Encode(responseData)
}

// Success Function
func Success(w http.ResponseWriter, message string) {
	if len(message) == 0 {
		message = "Success"
	}
	resp := ResSuccess{Status: true, Code: http.StatusOK, Message: message}
	Write(w, resp.Code, resp)
}

// Created Function
func Created(w http.ResponseWriter) {
	resp := ResSuccess{Status: true, Code: http.StatusCreated, Message: "Created"}
	Write(w, resp.Code, resp)
}

// Updated Function
func Updated(w http.ResponseWriter) {
	resp := ResSuccess{Status: true, Code: http.StatusOK, Message: "Updated"}
	Write(w, resp.Code, resp)
}

// NoContent Function
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(204)
}

// NotFound Function
func NotFound(w http.ResponseWriter, message string) {
	if len(message) == 0 {
		message = "Not Found"
	}
	resp := ResError{Status: false, Code: http.StatusNotFound, Message: "Not Found", Error: message}
	Write(w, resp.Code, resp)
}

// Unauthorized Function
func Unauthorized(w http.ResponseWriter) {
	resp := ResError{Status: false, Code: http.StatusUnauthorized, Message: "Unauthorized", Error: "Unauthorized"}
	Write(w, resp.Code, resp)
}

// Authenticate Function
func Authenticate(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Authorization Required"`)
	Unauthorized(w)
}

// InternalError Function
func InternalError(w http.ResponseWriter, message string) {
	if len(message) == 0 {
		message = "Internal Server Error"
	}
	resp := ResError{Status: false, Code: http.StatusInternalServerError, Message: "Internal Server Error", Error: message}
	logger.Println(logger.LogLevelError, "http-access", strings.ToLower(message))
	Write(w, resp.Code, resp)
}

// BadRequest Function
func BadRequest(w http.ResponseWriter, message string) {
	if len(message) == 0 {
		message = "Bad Request"
	}
	resp := ResError{Status: false, Code: http.StatusBadRequest, Message: "Bad Request", Error: message}
	logger.Println(logger.LogLevelError, "http-access", strings.ToLower(message))
	Write(w, resp.Code, resp)
}

// BadGateway Function
func BadGateway(w http.ResponseWriter, message string) {
	if len(message) == 0 {
		message = "Bad Gateway"
	}
	resp := ResError{Status: false, Code: http.StatusBadGateway, Message: "Bad Gateway", Error: message}
	logger.Println(logger.LogLevelError, "http-access", strings.ToLower(message))
	Write(w, resp.Code, resp)
}

// MethodNotAllowed Function
func MethodNotAllowed(w http.ResponseWriter, message string) {
	if len(message) == 0 {
		message = "Method Not Allowed"
	}
	resp := ResError{Status: false, Code: http.StatusMethodNotAllowed, Message: "Method Not Allowed", Error: message}
	Write(w, resp.Code, resp)
}
