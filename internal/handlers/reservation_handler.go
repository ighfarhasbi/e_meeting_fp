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
// @Success 200 {object} utils.SuccessResponse{data=models.PersonalDataCalculation}
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

	// masukkan rooms
	calculation.Rooms = append(calculation.Rooms, models.RoomCalculation{
		Name:          "Room 1",
		Type:          "Meeting Room",
		PricePerHour:  100000,
		Capacity:      10,
		ImgPath:       "https://example.com/room1.jpg",
		SubTotalSnack: 400000,
		SubTotalRoom:  100000,
		Snacks: models.SnacksCalculation{
			ID:       1,
			Category: "Snack",
			Name:     "Snack 1",
			Price:    100000,
		},
	})
	calculation.Rooms = append(calculation.Rooms, models.RoomCalculation{
		Name:          "Room 2",
		Type:          "Meeting Room",
		PricePerHour:  100000,
		Capacity:      10,
		ImgPath:       "https://example.com/room2.jpg",
		SubTotalSnack: 400000,
		SubTotalRoom:  100000,
		Snacks: models.SnacksCalculation{
			ID:       2,
			Category: "Snack",
			Name:     "Snack 2",
			Price:    100000,
		},
	})
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Success",
		Data:    calculation,
	})
}
