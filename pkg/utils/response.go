package utils

type SuccessResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type RegisterResposnse struct {
	Message string `json:"message"`
}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type MultupleErrorResponse struct {
	Errors []string `json:"massage"`
}

type ListResponse struct {
	Message   string `json:"message"`
	Data      any    `json:"data"`
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
	TotalPage int    `json:"totalPage"`
	TotalData int    `json:"totalData"`
}
