package entity

type PasswordResets struct {
	ID    int    `json:"id" db:"id"`
	Token string `json:"token" db:"token"`
}
