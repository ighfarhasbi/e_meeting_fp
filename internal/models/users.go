package models

type Users struct {
	ID        int    `json:"id" db:"users_id"`
	Username  string `json:"username" db:"username"`
	Email     string `json:"email" db:"email"`
	Role      string `json:"role" db:"role"`
	Status    string `json:"status" db:"status"`
	Language  string `json:"language" db:"language"`
	ImgPath   string `json:"img_path" db:"img_path"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}

type RegisterUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
