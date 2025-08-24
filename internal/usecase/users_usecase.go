package usecase

import (
	"e_meeting/internal/entity"
	"e_meeting/internal/repository"
	"e_meeting/pkg/utils"
	"errors"
)

type UserUsecase struct {
	repo repository.UsersRepository
}

// Membuat instance UserUsecase
func NewUserUsecase(r repository.UsersRepository) *UserUsecase {
	return &UserUsecase{repo: r}
}

func (uc *UserUsecase) Register(user *entity.Users, confirmPass string) error {
	// validasi email
	if err := utils.ValidateEmail(user.Email); err != nil {
		return errors.New("invalid email : " + err.Error())
	}
	// validasi confirm password
	if user.Password != confirmPass {
		return errors.New("password and confirm password do not match")
	}
	// validasi password characters
	if err := utils.ValidatePasswordCharacters(user.Password); err != nil {
		return errors.New("password validation failed : " + err.Error())
	}
	// hash password
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return errors.New("failed to hash password : " + err.Error())
	}
	user.Password = hashedPassword

	return uc.repo.Register(user)
}

func (uc *UserUsecase) Login(username, password string) (string, string, error) {
	u, err := uc.repo.FindByUsername(username)
	if err != nil {
		return "", "", errors.New("invalid username or password")
	}
	// validasi password
	if err := utils.ValidatePassword(u.Password, password); err != nil {
		return "", "", errors.New("invalid username or password")
	}
	// generate JWT token
	accessToken, refreshToken, err := utils.GenerateJWTToken(u.ID, u.Username, u.Role, u.Status)
	if err != nil {
		return "", "", errors.New("failed to generate JWT token : " + err.Error())
	}
	return accessToken, refreshToken, nil
}
