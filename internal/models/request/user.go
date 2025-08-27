package request

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	ConfirmPass string `json:"confirmPassword"`
}

type CheckEmailRequest struct {
	Email string `json:"email"`
}

type UpdatePassRequest struct {
	NewPassword    string `json:"newPassword"`
	NewConfirmPass string `json:"newConfirmPassword"`
}
