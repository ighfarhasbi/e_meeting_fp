package utils

import (
	"e_meeting/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWTToken(username string, role string, status string) (string, string, error) {
	cfg := config.New() // Ambil JWT secret dari konfigurasi
	if cfg.JWTSecret == "" {
		return "", "", jwt.ErrInvalidKey
	}
	// Generate token jwt dengan klaim yang sesuai
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"role":     role,
		"status":   status,
		"type":     "access",
		"exp":      time.Now().Add(24 * time.Hour).Unix(), // Token berlaku selama 24 jam
		"iat":      time.Now().Unix(),
	})
	accessTokenStr, err := accessToken.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"role":     role,
		"status":   status,
		"type":     "refresh",
		"exp":      time.Now().Add(7 * 24 * time.Hour).Unix(), // Token berlaku selama 7 hari
		"iat":      time.Now().Unix(),
	})
	refreshTokenStr, err := refreshToken.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenStr, refreshTokenStr, nil
}
