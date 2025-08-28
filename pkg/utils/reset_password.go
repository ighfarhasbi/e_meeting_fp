package utils

import (
	"e_meeting/config"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gopkg.in/gomail.v2"
)

func SendEmail(to, token string) error {
	cfg := config.New()
	resetURL := fmt.Sprintf("http://localhost:8080/reset_password/%s", token) // frontend link

	m := gomail.NewMessage()
	m.SetHeader("From", "ighfarhasbiash@gmail.com")
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Reset Your Password")
	m.SetBody("text/html", fmt.Sprintf(`
        <p>Click the link below to reset your password:</p>
        <a href="%s">Reset Password</a>
    `, resetURL))

	d := gomail.NewDialer("smtp.gmail.com", 587, "ighfarhasbiash@gmail.com", cfg.GmailPassword)
	return d.DialAndSend(m)
}

func GenerateResetToken(email string, id int) (string, error) {
	cfg := config.New()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email":   email,
		"user_id": id,
		"type":    "reset_password",
		"exp":     time.Now().Add(time.Duration(cfg.ExpRefreshToken) * time.Second).Unix(), // Token berlaku selama 15 menit
		"iat":     time.Now().Unix(),
	})
	tokenStr, err := token.SignedString([]byte(cfg.JWTResetPassword))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenStr, nil
}
