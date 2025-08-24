package response

type LoginResponse struct {
	Message string        `json:"message"`
	Data    TokenResponse `json:"data"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
