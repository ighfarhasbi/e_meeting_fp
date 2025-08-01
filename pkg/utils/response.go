package utils

type Response struct {
	Data    any    `json:"data"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
