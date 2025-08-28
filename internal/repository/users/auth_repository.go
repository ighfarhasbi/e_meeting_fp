package repository

import "e_meeting/internal/entity"

// mendefinisikan method apa saja yang akan digunakan dan berhubungan langsung hanya dengan database
type AuthRepository interface {
	InsertDataUser(user *entity.Users) error
	FindByUsername(username string) (*entity.Users, error)
}
