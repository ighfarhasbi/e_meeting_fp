package usecase

import (
	"e_meeting/internal/models/request"
	repository "e_meeting/internal/repository/users"
	"e_meeting/pkg/utils"
	"errors"
	"log"

	"github.com/golang-jwt/jwt/v5"
)

type ResetPassUsecase struct {
	repo repository.ResetPassRepository
}

func NewResetPassUsecase(r repository.ResetPassRepository) *ResetPassUsecase {
	return &ResetPassUsecase{repo: r}
}

func (uc *ResetPassUsecase) CheckEmailExists(email string) error {
	// panggil repository untuk memeriksa apakah email sudah ada di database
	u, err := uc.repo.GetUserByEmail(email)
	if err != nil {
		return err
	}
	// generate token reset password dengan claim email dan userID
	token, err := utils.GenerateResetToken(u.Email, u.ID)
	if err != nil {
		return err
	}

	// kirim email reset password ke user dengan goroutine
	go func(email, token string) {
		if err := utils.SendEmail(email, token); err != nil {
			log.Printf("[SendEmail Error] to %s: %v", email, err)
		}
	}(u.Email, token)

	// panggil repository untuk menyimpan token reset password ke database
	if err := uc.repo.InsertResetToken(token); err != nil {
		return err
	}
	return nil
}

func (uc *ResetPassUsecase) ResetPassword(token, secretKey string, req request.UpdatePassRequest) error {
	// panggil repository untuk ambil token
	t, err := uc.repo.GetToken(token)
	if err != nil {
		return err
	}
	// parse token
	parsedToken, err := jwt.Parse(t.Token, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})
	// validasi token
	if err != nil || !parsedToken.Valid {
		return errors.New("invalid reset token")
	}
	// ambil userID dari claim token
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		return errors.New("invalid token claims")
	}
	userID := int(claims["user_id"].(float64))

	// validasi password characters
	if err := utils.ValidatePasswordCharacters(req.NewPassword); err != nil {
		return err
	}
	// validasi confirm password
	if req.NewPassword != req.NewConfirmPass {
		return errors.New("password and confirm password do not match")
	}
	// hash password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	// begin transaction
	tx, err := uc.repo.BeginTx()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// panggil repository untuk update password user
	if err := uc.repo.UpdatePassword(tx, userID, hashedPassword); err != nil {
		return err
	}
	// panggil repository untuk menghapus token reset password
	if err := uc.repo.DeleteToken(tx, t.ID); err != nil {
		return err
	}

	// jika tidak ada error, commit transaction
	return tx.Commit()
}
