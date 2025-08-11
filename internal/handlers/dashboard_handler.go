package handlers

import (
	"database/sql"
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func InitDashboardHandler(e *echo.Group, dbConn *sql.DB) {
	e.GET("/dashboard", func(c echo.Context) error {
		return GetDashboardCalculation(c, dbConn)
	})
}

// @Summary Get dashboard calculation
// @Description Get dashboard calculation
// @Tags Dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=models.Dashboard}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /dashboard [get]
func GetDashboardCalculation(c echo.Context, db *sql.DB) error {
	// ambil claim token dari context
	claims := c.Get("client").(jwt.MapClaims)

	// ambil role dari klaim token
	role, ok := claims["role"].(string)
	// jika role bukan admin, kembalikan response unauthorized
	if !ok || role != "admin" {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Unauthorized access",
		})
	}

	// hitung total omzet
	var totalOmzet float64
	err := db.QueryRow("SELECT SUM(total) FROM transactions WHERE status = 'paid'::tx_status_enum").Scan(&totalOmzet)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Message: "Failed to count transactions: " + err.Error()})
	}

	// hitung total transaksi
	var totalTransactions int
	err = db.QueryRow("SELECT COUNT(*) FROM transactions WHERE status = 'paid'::tx_status_enum").Scan(&totalTransactions)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Message: "Failed to count transactions: " + err.Error()})
	}

	// hitung total room
	var totalRooms int
	err = db.QueryRow("SELECT COUNT(*) FROM rooms").Scan(&totalRooms)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Message: "Failed to count rooms: " + err.Error()})
	}

	// hitung total participant di tabel detail_transaction yang statusnya sudah paid di tabel transactions
	var totalParticipants int
	err = db.QueryRow("SELECT SUM(participants) FROM detail_transaction dt JOIN transactions t ON dt.tx_id = t.tx_id WHERE t.status = 'paid'::tx_status_enum").Scan(&totalParticipants)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Message: "Failed to count participants: " + err.Error()})
	}

	// hitung total transaksi semua room
	var totalTransactionsAll int
	err = db.QueryRow(`
            SELECT COUNT(*)
            FROM detail_transaction dt
            JOIN transactions t ON dt.tx_id = t.tx_id
            WHERE t.status = 'paid'::tx_status_enum
        `).Scan(&totalTransactionsAll)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to count all transactions: " + err.Error(),
		})
	}

	// ambil data per room (jumlah transaksi dan omzet)
	rows, err := db.Query(`
            SELECT dt.rooms_id, r.name, SUM(t.total) AS omzet, COUNT(*) AS trx_count
            FROM detail_transaction dt
            JOIN rooms r ON r.rooms_id = dt.rooms_id
            JOIN transactions t ON t.tx_id = dt.tx_id
            WHERE t.status = 'paid'::tx_status_enum
            GROUP BY dt.rooms_id, r.name
        `)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to get rooms: " + err.Error(),
		})
	}
	defer rows.Close()

	var rooms []models.RoomsDashboard
	for rows.Next() {
		var room models.RoomsDashboard
		var trxCount int

		if err := rows.Scan(&room.ID, &room.Name, &room.Omzet, &trxCount); err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Failed to scan room: " + err.Error(),
			})
		}

		// hitung persentase penggunaan berdasarkan jumlah transaksi
		if totalTransactionsAll > 0 {
			room.PercentegeOfUsage = int(float64(trxCount) / float64(totalTransactionsAll) * 100)
		}

		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to iterate rooms: " + err.Error(),
		})
	}

	result := models.Dashboard{
		TotalRoom:        totalRooms,
		TotalVisitor:     totalParticipants,
		TotalReservation: totalTransactions,
		TotalOmzet:       totalOmzet,
		Rooms:            rooms,
	}

	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Success",
		Data:    result,
	})
}
