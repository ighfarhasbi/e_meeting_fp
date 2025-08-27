package repository

import "e_meeting/internal/entity"

type AuthRepository interface {
	InsertDataUser(user *entity.Users) error
	FindByUsername(username string) (*entity.Users, error)
}
