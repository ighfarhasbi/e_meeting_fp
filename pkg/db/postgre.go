package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// NewPostgres membuat koneksi ke PostgreSQL dan mengembalikan *sqlx.DB
func NewPostgres(dbUrl string) (*sql.DB, error) {
	// db, err := sqlx.Connect("postgres", dbUrl)
	// if err != nil {
	// 	return nil, err
	// }

	// // Set konfigurasi pool connection
	// db.SetMaxOpenConns(25)                 // Maksimal koneksi terbuka
	// db.SetMaxIdleConns(25)                 // Maksimal koneksi idle
	// db.SetConnMaxLifetime(5 * time.Minute) // Maksimal lifetime koneksi

	// log.Println("PostgreSQL connected successfully")
	// return db, nil

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Set konfigurasi pool connection
	db.SetMaxOpenConns(25)                  // maksimal 25 koneksi aktif
	db.SetMaxIdleConns(10)                  // simpan 10 koneksi idle
	db.SetConnMaxLifetime(5 * time.Minute)  // koneksi akan diganti tiap 5 menit
	db.SetConnMaxIdleTime(30 * time.Second) // koneksi idle akan ditutup setelah 30 detik

	// Test koneksi
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error ping database: %w", err)
	}

	log.Println("Connected to PostgreSQL")
	return db, nil
}
