package repository

import "e_meeting/internal/entity"

type UsersRepository interface {
	Register(user *entity.Users) error
	FindByUsername(username string) (*entity.Users, error)
}
