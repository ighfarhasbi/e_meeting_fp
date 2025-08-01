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
