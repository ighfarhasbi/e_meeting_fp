package handlers

import (
	"database/sql"
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

func InitRoomHandler(e *echo.Group, dbConn *sql.DB) {
	e.GET("/rooms", func(c echo.Context) error {
		return GetRooms(c, dbConn)
	})
}

func GetRooms(c echo.Context, db *sql.DB) error {
	// ambil daftar ruangan dari database
	rows, err := db.Query("SELECT rooms_id, name, type, price_perhour, capacity, img_path, created_at, updated_at FROM rooms")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Gagal mengambil daftar rooms",
		})
	}
	defer rows.Close()

	// buat slice untuk menyimpan daftar rooms
	var rooms []models.Room
	// iterasi melalui hasil query dan masukkan ke dalam slice
	for rows.Next() {
		var room models.Room
		if err := rows.Scan(&room.ID, &room.Name, &room.Type, &room.PricePerHour, &room.Capacity, &room.ImgPath, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Gagal mengambil daftar rooms",
			})
		}
		rooms = append(rooms, room)
	}
	// periksa apakah ada error saat iterasi
	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Gagal mengambil daftar rooms",
		})
	}

	// kembalikan daftar rooms sebagai response
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Daftar rooms berhasil diambil",
		Data:    rooms,
	})
}
