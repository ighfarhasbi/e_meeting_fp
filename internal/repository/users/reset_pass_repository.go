package repository

import (
	"database/sql"
	"e_meeting/internal/entity"
)

type ResetPassRepository interface {
	// untuk usecase CheckEmailExists
	GetUserByEmail(email string) (*entity.Users, error)
	InsertResetToken(token string) error
	// untuk usecase ResetPassword
	GetToken(token string) (*entity.PasswordResets, error)
	BeginTx() (*sql.Tx, error)
	UpdatePassword(tx *sql.Tx, userID int, newPassword string) error
	DeleteToken(tx *sql.Tx, tokenID int) error
}
