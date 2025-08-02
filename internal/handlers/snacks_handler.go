package handlers

import (
	"database/sql"
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

// InitSnacksHandler initializes the snacks handler
func InitSnacksHandler(e *echo.Group, dbConn *sql.DB) {
	e.GET("/snacks", func(c echo.Context) error {
		return GetSnacks(c, dbConn)
	})
}

func GetSnacks(c echo.Context, db *sql.DB) error {
	// ambil daftar snacks dari database
	rows, err := db.Query("SELECT snacks_id, name, price, category FROM snacks")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Gagal mengambil daftar snacks",
		})
	}
	defer rows.Close()

	// buat slice untuk menyimpan daftar snacks
	var snacks []models.Snacks
	// iterasi melalui hasil query dan masukkan ke dalam slice
	for rows.Next() {
		var snack models.Snacks
		if err := rows.Scan(&snack.ID, &snack.Name, &snack.Price, &snack.Category); err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Gagal mengambil daftar snacks",
			})
		}
		snacks = append(snacks, snack)
	}
	// periksa apakah ada error saat iterasi
	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Gagal mengambil daftar snacks",
		})
	}

	// kembalikan daftar snacks sebagai response
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Daftar snacks berhasil diambil",
		Data:    snacks,
	})
}
