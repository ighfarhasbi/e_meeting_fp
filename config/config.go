package config

import (
	"log"
	"os"
	"strconv"
)

// Config menyimpan konfigurasi aplikasi
type Config struct {
	DBUrl            string
	JWTSecret        string
	Port             string
	GmailPassword    string // untuk mengirim email
	JWTResetPassword string // untuk reset password
	Domain           string // domain untuk url image
	RedisUrl         string // untuk redis
	ExpAccessToken   int    // in seconds
	ExpRefreshToken  int    // in seconds
	ExpResetPassword int    // in seconds
}

// New membaca konfigurasi dari environment
func New() *Config {
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL is not set in env")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set in env")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default
	}

	gmailPassword := os.Getenv("GMAIL_PASSWORD")
	if gmailPassword == "" {
		log.Fatal("GMAIL_PASSWORD is not set in env")
	}

	jwtResetPassword := os.Getenv("JWT_RESET_PASSWORD")
	if jwtResetPassword == "" {
		log.Fatal("JWT_RESET_PASSWORD is not set in env")
	}

	domain := os.Getenv("DOMAIN")
	if domain == "" {
		log.Fatal("DOMAIN is not set in env")
	}

	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		log.Fatal("REDIS_URL is not set in env")
	}

	var (
		expAccessToken   int
		expRefreshToken  int
		expResetPassword int
	)
	if os.Getenv("EXP_ACCESS_TOKEN") != "" {
		var err error
		expAccessToken, err = strconv.Atoi(os.Getenv("EXP_ACCESS_TOKEN"))
		if err != nil {
			log.Fatal("Invalid EXP_ACCESS_TOKEN in env")
		}
	}
	if os.Getenv("EXP_REFRESH_TOKEN") != "" {
		var err error
		expRefreshToken, err = strconv.Atoi(os.Getenv("EXP_REFRESH_TOKEN"))
		if err != nil {
			log.Fatal("Invalid EXP_REFRESH_TOKEN in env")
		}
	}
	if os.Getenv("EXP_RESET_PASSWORD_TOKEN") != "" {
		var err error
		expResetPassword, err = strconv.Atoi(os.Getenv("EXP_RESET_PASSWORD_TOKEN"))
		if err != nil {
			log.Fatal("Invalid EXP_RESET_PASSWORD_TOKEN in env")
		}
	}

	return &Config{
		DBUrl:            dbUrl,
		JWTSecret:        jwtSecret,
		Port:             port,
		GmailPassword:    gmailPassword,
		JWTResetPassword: jwtResetPassword,
		Domain:           domain,
		RedisUrl:         redisUrl,
		ExpAccessToken:   expAccessToken,
		ExpRefreshToken:  expRefreshToken,
		ExpResetPassword: expResetPassword,
	}
}
