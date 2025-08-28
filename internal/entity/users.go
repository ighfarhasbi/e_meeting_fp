package entity

import "time"

type Users struct {
	ID        int       `json:"id" db:"users_id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Role      string    `json:"role" db:"role"`
	Password  string    `json:"password" db:"password"`
	Status    string    `json:"status" db:"status"`
	Language  string    `json:"language" db:"language"`
	ImgUrl    string    `json:"imgUrl" db:"img_path"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}
