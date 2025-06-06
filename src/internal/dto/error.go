package dto

type SuccessResponse struct {
	Code    *int   `json:"code,omitempty"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Code  *int   `json:"code,omitempty"`
	Error string `json:"error"`
}
