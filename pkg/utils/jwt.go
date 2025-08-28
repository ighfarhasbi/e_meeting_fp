package utils

import (
	"e_meeting/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWTToken(id int, username string, role string, status string) (string, string, error) {
	cfg := config.New() // Ambil JWT secret dari konfigurasi
	if cfg.JWTSecret == "" {
		return "", "", jwt.ErrInvalidKey
	}
	// Generate token jwt dengan klaim yang sesuai
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       id,
		"username": username,
		"role":     role,
		"status":   status,
		"type":     "access",
		"exp":      time.Now().Add(time.Duration(cfg.ExpAccessToken) * time.Second).Unix(), // Token berlaku selama 24 jam
		"iat":      time.Now().Unix(),
	})
	accessTokenStr, err := accessToken.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       id,
		"username": username,
		"role":     role,
		"status":   status,
		"type":     "refresh",
		"exp":      time.Now().Add(time.Duration(cfg.ExpRefreshToken) * time.Second).Unix(), // Token berlaku selama 7 hari
		"iat":      time.Now().Unix(),
	})
	refreshTokenStr, err := refreshToken.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenStr, refreshTokenStr, nil
}

func RefreshAccessToken(refreshToken string) (string, error) {
	cfg := config.New() // Ambil JWT secret dari konfigurasi
	if cfg.JWTSecret == "" {
		return "", jwt.ErrInvalidKey
	}

	// Parse refresh token
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", jwt.ErrInvalidKey
	}

	// Periksa apakah token adalah refresh token
	if claims, ok := token.Claims.(jwt.MapClaims); !ok || claims["type"] != "refresh" {
		return "", jwt.ErrInvalidType
	}

	// Generate new access token
	id := token.Claims.(jwt.MapClaims)["id"].(float64)
	username := token.Claims.(jwt.MapClaims)["username"].(string)
	role := token.Claims.(jwt.MapClaims)["role"].(string)
	status := token.Claims.(jwt.MapClaims)["status"].(string)

	// call GenerateJWTToken to create a new access token (and optionally a new refresh token)
	// newAccessToken, _, err := GenerateJWTToken(username, role, status)
	// if err != nil {
	// 	return "", err
	// }

	// hanya buat access token baru
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       id,
		"username": username,
		"role":     role,
		"status":   status,
		"type":     "access",
		"exp":      time.Now().Add(time.Duration(cfg.ExpAccessToken) * time.Second).Unix(), // Token berlaku selama 24 jam
		"iat":      time.Now().Unix(),
	})
	accessTokenStr, err := accessToken.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return accessTokenStr, nil
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	cfg := config.New() // Ambil JWT secret dari konfigurasi
	if cfg.JWTSecret == "" {
		return nil, jwt.ErrInvalidKey
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	// Periksa apakah token adalah access token
	if claims, ok := token.Claims.(jwt.MapClaims); !ok || claims["type"] != "access" {
		return nil, jwt.ErrInvalidType
	}

	return token, nil
}
