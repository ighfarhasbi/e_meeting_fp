package config

import (
	"log"
	"os"
)

// Config menyimpan konfigurasi aplikasi
type Config struct {
	DBUrl     string
	JWTSecret string
	Port      string
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

	return &Config{
		DBUrl:     dbUrl,
		JWTSecret: jwtSecret,
		Port:      port,
	}
}
