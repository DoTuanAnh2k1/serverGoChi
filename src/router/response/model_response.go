package response

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
