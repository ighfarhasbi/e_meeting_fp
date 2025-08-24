package repository

import (
	"database/sql"
	"e_meeting/internal/entity"
)

type DBUsersRepository struct {
	DB *sql.DB
}

func NewDBUsersRepository(db *sql.DB) *DBUsersRepository {
	return &DBUsersRepository{DB: db}
}

func (r *DBUsersRepository) Register(user *entity.Users) error {
	_, err := r.DB.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)",
		user.Username, user.Email, user.Password)
	return err
}

func (r *DBUsersRepository) FindByUsername(username string) (*entity.Users, error) {
	row := r.DB.QueryRow("SELECT users_id, username, email, password, role, status, language, img_path, created_at, updated_at FROM users WHERE username = $1", username)
	var u entity.Users
	if err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.Status, &u.Language, &u.ImgUrl, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}
