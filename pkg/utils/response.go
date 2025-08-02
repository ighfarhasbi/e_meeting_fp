package utils

type Response struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type RegisterResposnse struct {
	Message string `json:"message"`
}

type LoginResponse struct {
	Message      string `json:"message"`
	Data         any    `json:"data"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
