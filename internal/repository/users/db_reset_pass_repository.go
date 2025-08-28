package repository

import (
	"database/sql"
	"e_meeting/internal/entity"
	"errors"
)

type DBResetPassRepository struct {
	DB *sql.DB
}

func NewDBResetPassRepository(db *sql.DB) *DBResetPassRepository {
	return &DBResetPassRepository{DB: db}
}

func (r *DBResetPassRepository) GetUserByEmail(email string) (*entity.Users, error) {
	// query untuk memeriksa apakah email sudah ada di database
	row := r.DB.QueryRow("SELECT users_id, username, email, role, status, language, img_path, created_at, updated_at FROM users WHERE email = $1", email)
	var u entity.Users
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Role, &u.Status, &u.Language, &u.ImgUrl, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("email not found: " + err.Error())
		}
		return nil, errors.New("failed to get user by email: " + err.Error())
	}
	return &u, nil
}

func (r *DBResetPassRepository) InsertResetToken(token string) error {
	_, err := r.DB.Exec("INSERT INTO password_resets (token) VALUES ($1)", token)
	if err != nil {
		return errors.New("failed to insert reset token: " + err.Error())
	}
	return nil
}

func (r *DBResetPassRepository) GetToken(token string) (*entity.PasswordResets, error) {
	// query untuk memeriksa apakah token sudah ada di database
	row := r.DB.QueryRow("SELECT id, token FROM password_resets WHERE token = $1", token)
	var t entity.PasswordResets
	if err := row.Scan(&t.ID, &t.Token); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("token not found: " + err.Error())
		}
		return nil, errors.New("failed to get token: " + err.Error())
	}
	return &t, nil
}

func (r *DBResetPassRepository) BeginTx() (*sql.Tx, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, errors.New("failed to begin transaction: " + err.Error())
	}
	return tx, nil
}
func (r *DBResetPassRepository) UpdatePassword(tx *sql.Tx, userID int, newPassword string) error {
	_, err := tx.Exec("UPDATE users SET password = $1 WHERE users_id = $2", newPassword, userID)
	if err != nil {
		return errors.New("failed to update password: " + err.Error())
	}
	return nil
}

func (r *DBResetPassRepository) DeleteToken(tx *sql.Tx, tokenID int) error {
	_, err := tx.Exec("DELETE FROM password_resets WHERE id = $1", tokenID)
	if err != nil {
		return errors.New("failed to delete token: " + err.Error())
	}
	return nil
}
