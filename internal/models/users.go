package models

type Users struct {
	ID        int    `json:"id" db:"users_id"`
	Username  string `json:"username" db:"username"`
	Email     string `json:"email" db:"email"`
	Role      string `json:"role" db:"role"`
	Status    string `json:"status" db:"status"`
	Language  string `json:"language" db:"language"`
	ImgPath   string `json:"imgUrl" db:"img_path"`
	CreatedAt string `json:"createdAt" db:"created_at"`
	UpdatedAt string `json:"updatedAt" db:"updated_at"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Language string `json:"language"`
	ImgPath  string `json:"imgUrl"`
	Password string `json:"password"`
}
