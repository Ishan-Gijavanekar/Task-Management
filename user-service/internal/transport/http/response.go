package http

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
