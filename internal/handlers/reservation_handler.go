package handlers

import (
	"database/sql"
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

func InitReservationHandler(e *echo.Group, dbConn *sql.DB) {
	e.GET("/reservations/calculation", func(c echo.Context) error {
		return ReservationCalculation(c, dbConn)
	})
}

// @Summary ReservationCalculation calculates the reservation
// @Tags Reservation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param room_id query string true "Room ID"
// @Param snack_id query string true "Snack ID"
// @Param start_time query string true "Start time"
// @Param end_time query string true "End time"
// @Param participants query string true "Participants"
// @Param user_id query string true "User ID"
// @Param name query string true "Name"
// @Param phone_number query string true "Phone number"
// @Param company query string true "Company"
// @Success 200 {object} utils.SuccessResponse{data=models.CalculationResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /reservations/calculation [get]
func ReservationCalculation(c echo.Context, db *sql.DB) error {
	// ambil query parameter
	roomID := c.QueryParam("room_id")
	snackID := c.QueryParam("snack_id")
	startTime := c.QueryParam("start_time")
	endTime := c.QueryParam("end_time")
	participants := c.QueryParam("participants")
	userID := c.QueryParam("user_id")
	name := c.QueryParam("name")
	phoneNumber := c.QueryParam("phone_number")
	company := c.QueryParam("company")

	// konversi string ke int
	roomIDInt, snackIDInt, participantsInt, userIDInt, err := utils.StringToInt(roomID, snackID, participants, userID)
	if err != nil || roomIDInt <= 0 || snackIDInt <= 0 || participantsInt <= 0 || userIDInt <= 0 {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid query parameters",
		})
	}

	// konversi startTime dan endTime ke time.Time
	startTimeTime, err := utils.StringToTimestamptz(startTime)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid start time",
		})
	}

	endTimeTime, err := utils.StringToTimestamptz(endTime)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid end time",
		})
	}
	duration := utils.CalculateDuration(startTimeTime, endTimeTime)

	// get room data dari database
	rows := db.QueryRow("SELECT name, type, price_perhour, capacity, img_path FROM rooms WHERE rooms_id = $1", roomIDInt)
	var room models.CURoomRequest
	err = rows.Scan(&room.Name, &room.Type, &room.PricePerHour, &room.Capacity, &room.ImgPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to get room data: " + err.Error(),
		})
	}
	// get snack data dari database
	rows = db.QueryRow("SELECT snacks_id, name, price, category FROM snacks WHERE snacks_id = $1", snackIDInt)
	var snack models.Snacks
	err = rows.Scan(&snack.ID, &snack.Name, &snack.Price, &snack.Category)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to get snack data: " + err.Error(),
		})
	}

	// validasi room capacity
	if room.Capacity < participantsInt {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Room capacity is not enough",
		})
	}

	// hitung subtotal snack
	snackPrice, _ := (snack.Price).Float64() // konversi ke float dari decimal.Decimal
	subTotalSnack := snackPrice * float64(participantsInt)
	// hitung subtotal room
	subTotalRoom := room.PricePerHour * float64(duration.Hours())

	// hitung total
	total := subTotalSnack + subTotalRoom

	// siapkan response
	var calculation models.CalculationResponse
	// masukkan personal data
	calculation.PersonalData.Name = name
	calculation.PersonalData.NoHp = phoneNumber
	calculation.PersonalData.Company = company
	calculation.PersonalData.StartTime = startTime
	calculation.PersonalData.EndTime = endTime
	calculation.PersonalData.Duration = int(duration.Hours())
	calculation.PersonalData.Participants = participantsInt
	calculation.Total = total

	// masukkan rooms
	calculation.Rooms = append(calculation.Rooms, models.RoomCalculation{
		Name:          room.Name,
		Type:          room.Type,
		PricePerHour:  room.PricePerHour,
		Capacity:      room.Capacity,
		ImgPath:       room.ImgPath,
		SubTotalSnack: subTotalSnack,
		SubTotalRoom:  subTotalRoom,
		Snacks: models.SnacksCalculation{
			ID:       snack.ID,
			Category: snack.Category,
			Name:     snack.Name,
			Price:    snackPrice,
		},
	})

	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Success",
		Data:    calculation,
	})
}
