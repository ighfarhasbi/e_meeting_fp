package handlers

import (
	"database/sql"
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

func InitReservationHandler(e *echo.Group, dbConn *sql.DB, rdb *redis.Client) {
	e.GET("/reservations/calculation", func(c echo.Context) error {
		return ReservationCalculation(c, dbConn)
	})
	e.POST("/reservations", func(c echo.Context) error {
		return CreateReservation(c, rdb)
	})
	e.GET("/reservations/history", func(c echo.Context) error {
		return HistoryReservations(c, dbConn)
	})
	e.GET("/reservations/:id", func(c echo.Context) error {
		return GetReservationByID(c, dbConn)
	})
	e.PATCH("/reservations/status/:id", func(c echo.Context) error {
		return UpdateReservationStatus(c, dbConn)
	})
}

// @Summary ReservationCalculation calculates the reservation
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param room_id query string true "Room ID"
// @Param snack_id query string false "Snack ID"
// @Param start_time query string true "Start time"
// @Param end_time query string true "End time"
// @Param participants query string true "Participants"
// @Param user_id query string true "User ID"
// @Param name query string true "Name"
// @Param phone_number query string true "Phone number"
// @Param company query string true "Company"
// @Success 200 {object} utils.SuccessResponse{data=models.CalculationResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
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
	roomIDInt, participantsInt, userIDInt, err := utils.StringToInt(roomID, participants, userID)
	if err != nil || roomIDInt <= 0 || participantsInt <= 0 || userIDInt <= 0 {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid query parameters",
		})
	}

	// ambil data startTime dan endTime di database untuk dibandingkan
	// bookTimes, err := db.Query("SELECT start_time, end_time FROM detail_transaction dt JOIN transactions t ON dt.tx_id = t.tx_id WHERE rooms_id = $1 AND t.status != 'canceled'", roomIDInt)
	// if err != nil {
	// 	return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
	// 		Message: "Failed to retrieve reservation: " + err.Error(),
	// 	})
	// }
	// var startTimeDB, endTimeDB []time.Time
	// for bookTimes.Next() {
	// 	var startTime, endTime time.Time
	// 	if err := bookTimes.Scan(&startTime, &endTime); err != nil {
	// 		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
	// 			Message: "Failed to retrieve reservation: " + err.Error(),
	// 		})
	// 	}
	// 	startTimeDB = append(startTimeDB, startTime)
	// 	endTimeDB = append(endTimeDB, endTime)
	// }

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

	// cek startTime dan endTime minimal hari ini dan waktu sekarang
	if startTimeTime.Before(time.Now()) || endTimeTime.Before(time.Now()) {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Reservation time must be after current time",
		})
	}

	// cek apakah startTime lebih besar atau sama dengan endTime
	if !startTimeTime.Before(endTimeTime) {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Start time must be before end time",
		})
	}

	// cek apakah startTime dan endTime beririsan dengan startTimeDB dan endTimeDB
	var overlaps bool
	err = db.QueryRow(`
    SELECT EXISTS (
        SELECT 1
        FROM detail_transaction dt
        JOIN transactions t ON dt.tx_id = t.tx_id
        WHERE rooms_id = $1
          AND t.status != 'canceled'
          AND (dt.start_time, dt.end_time) OVERLAPS ($2, $3)
    )
	`, roomIDInt, startTimeTime, endTimeTime).Scan(&overlaps)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to check overlap: " + err.Error(),
		})
	}

	if overlaps {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Reservation time overlaps with existing reservation",
		})
	}

	// // cek apakah interval startTime dan endTime beririsan dengan startTimeDB dan endTimeDB
	// if utils.IsOverlapping(startTimeTime, endTimeTime, startTimeDB, endTimeDB) {
	// 	return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
	// 		Message: "Reservation time overlaps with existing reservation",
	// 	})
	// }

	// hitung durasi
	duration := utils.CalculateDuration(startTimeTime, endTimeTime)

	// get room data dari database
	rows := db.QueryRow("SELECT name, type, price_perhour, capacity, img_path FROM rooms WHERE rooms_id = $1", roomIDInt)
	var room models.CURoomRequest
	if err := rows.Scan(&room.Name, &room.Type, &room.PricePerHour, &room.Capacity, &room.ImgPath); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Message: "Room not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to get room data: " + err.Error(),
		})
	}
	// get snack data dari database jika snack_id ada
	var snackIDInt int
	var snack models.Snacks
	if snackID != "" {
		snackIDInt, err = strconv.Atoi(snackID)
		if err != nil || snackIDInt <= 0 {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Message: "Invalid snack ID",
			})
		}
		rows := db.QueryRow("SELECT snacks_id, name, price, category FROM snacks WHERE snacks_id = $1", snackIDInt)
		if err := rows.Scan(&snack.ID, &snack.Name, &snack.Price, &snack.Category); err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusNotFound, utils.ErrorResponse{
					Message: "Snack not found",
				})
			}
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Failed to get snack data: " + err.Error(),
			})
		}
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

// Summary CreateReservation creates a new reservation
// Description Create a new reservation with the provided details
// Tags reservations
// Accept json
// Produce json
// Security BearerAuth
// Param request body models.ReservationRequest true "Reservation details"
// Success 201 {object} utils.SuccessResponse{data=nil}
// Sucess 202 {object} utils.SuccessResponse{data=nil}
// Failure 400 {object} utils.ErrorResponse
// Failure 401 {object} utils.ErrorResponse
// Failure 500 {object} utils.ErrorResponse
// Router /reservations [post]
// func CreateReservation(c echo.Context, db *sql.DB) error {
// 	// ambil claim token
// 	claims := c.Get("client").(jwt.MapClaims)
// 	// ambil id user
// 	userIDfloat, ok := claims["id"].(float64)
// 	if !ok {
// 		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
// 			Message: "Invalid token claims",
// 		})
// 	}
// 	userIDInt := int(userIDfloat) // konversi ke int

// 	// ambil body request
// 	var request models.ReservationRequest
// 	if err := c.Bind(&request); err != nil {
// 		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
// 			Message: "Invalid request body",
// 		})
// 	}

// 	// validasi data
// 	if request.UserID != userIDInt {
// 		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
// 			Message: "Invalid user ID",
// 		})
// 	}

// 	// mulai tx dari ambil data sampai transaksi selesai
// 	tx, err := db.Begin()
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
// 			Message: "Failed to begin transaction: " + err.Error(),
// 		})
// 	}
// 	defer tx.Rollback()

// 	// ambil data snack dari database jika snack_id ada
// 	var snackReq any
// 	snackReq = request.Rooms[0].SnackID
// 	fmt.Println("id snack: ", snackReq)
// 	var snack models.Snacks
// 	if snackReq != 0 {
// 		row := tx.QueryRow("SELECT snacks_id, name, price, category FROM snacks WHERE snacks_id = $1", snackReq)
// 		if err := row.Scan(&snack.ID, &snack.Name, &snack.Price, &snack.Category); err != nil {
// 			if err == sql.ErrNoRows {
// 				return c.JSON(http.StatusNotFound, utils.ErrorResponse{
// 					Message: "Snack not found",
// 				})
// 			}
// 			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
// 				Message: "Failed to get snack data: " + err.Error(),
// 			})
// 		}
// 	} else if snackReq == 0 {
// 		snackReq = nil // set nil karena di db nullable
// 	}

// 	// ambil data room dari database
// 	var room models.CURoomRequest
// 	row := tx.QueryRow("SELECT name, type, price_perhour, capacity, img_path FROM rooms WHERE rooms_id = $1", request.Rooms[0].ID)
// 	if err := row.Scan(&room.Name, &room.Type, &room.PricePerHour, &room.Capacity, &room.ImgPath); err != nil {
// 		if err == sql.ErrNoRows {
// 			return c.JSON(http.StatusNotFound, utils.ErrorResponse{
// 				Message: "Room not found",
// 			})
// 		}
// 		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
// 			Message: "Failed to get room data: " + err.Error(),
// 		})
// 	}

// 	// hitung durasi
// 	startTimeTime, err := utils.StringToTimestamptz(request.Rooms[0].StartTime)
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
// 			Message: "Invalid start time format",
// 		})
// 	}
// 	endTimeTime, err := utils.StringToTimestamptz(request.Rooms[0].EndTime)
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
// 			Message: "Invalid end time format",
// 		})
// 	}
// 	duration := utils.CalculateDuration(startTimeTime, endTimeTime)

// 	// hitung subtotal
// 	snackPrice, _ := (snack.Price).Float64()
// 	subTotalSnack := snackPrice * float64(request.Rooms[0].Participants)
// 	subTotalRoom := room.PricePerHour * float64(duration.Hours())
// 	total := subTotalSnack + subTotalRoom

// 	// uuid generate
// 	uuidTx := uuid.New().String()       // tx_id
// 	uuidDetailTx := uuid.New().String() // detail_tx_id

// 	// insert data ke tabel transaction
// 	_, err = tx.Exec("INSERT INTO transactions (tx_id, users_id, name, no_hp, company, note, total) VALUES ($1, $2, $3, $4, $5, $6, $7)",
// 		uuidTx, request.UserID, request.Name, request.PhoneNumber, request.Company, request.Notes, total)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
// 			Message: "Failed to insert transaction data: " + err.Error(),
// 		})
// 	}
// 	// insert data ke tabel detail_transaction
// 	_, err = tx.Exec("INSERT INTO detail_transaction (detail_tx_id, tx_id, rooms_id, start_time, end_time, participants, snacks_id, sub_total_snacks, sub_total_price_room, price_snack_perpack, price_room_perhour) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10 , $11)",
// 		uuidDetailTx, uuidTx, request.Rooms[0].ID, startTimeTime, endTimeTime, request.Rooms[0].Participants, snackReq, subTotalSnack, subTotalRoom, snack.Price, room.PricePerHour)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
// 			Message: "Failed to insert detail transaction data: " + err.Error(),
// 		})
// 	}
// 	// commit tx
// 	if err := tx.Commit(); err != nil {
// 		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
// 			Message: "Failed to commit transaction: " + err.Error(),
// 		})
// 	}

// 	return c.JSON(http.StatusCreated, utils.SuccessResponse{
// 		Message: "Success",
// 		Data:    nil,
// 	})
// }

// @Summary HistoryReservations gets history reservation
// @Description HistoryReservations gets history reservation
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param startDate query string false "Start date"
// @Param endDate query string false "End date"
// @Param status query string false "Status"
// @Param type query string false "Type"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {object} utils.ListResponse{data=[]models.TransactionResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /reservations/history [get]
func HistoryReservations(c echo.Context, db *sql.DB) error {
	// Ambil query parameter untuk filter
	startDateStr := c.QueryParam("startDate")
	endDateStr := c.QueryParam("endDate")
	status := c.QueryParam("status")
	typeStr := c.QueryParam("type")

	// Pagination params
	pageStr := c.QueryParam("page")
	pageSizeStr := c.QueryParam("pageSize")

	var page, pageSize int
	var err error

	if pageStr != "" {
		if page, err = strconv.Atoi(pageStr); err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Message: "Invalid page parameter"})
		}
	} else {
		page = 1
	}
	if pageSizeStr != "" {
		if pageSize, err = strconv.Atoi(pageSizeStr); err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{Message: "Invalid pageSize parameter"})
		}
	} else {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// Ambil user_id dari token
	claims := c.Get("client").(jwt.MapClaims)
	userIDfloat, ok := claims["id"].(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{Message: "Invalid token claims"})
	}
	userID := int(userIDfloat)
	// ambil role admin dari klaim token
	role, ok := claims["role"].(string)
	// jika role bukan admin, kembalikan response unauthorized
	if !ok {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Unauthorized access",
		})
	}

	// Hitung total data
	var totalData int
	// validasi role
	switch role {
	case "user":
		countQuery := `
        SELECT COUNT(DISTINCT t.tx_id)
        FROM transactions t
        JOIN detail_transaction dt ON t.tx_id = dt.tx_id
        JOIN rooms r ON dt.rooms_id = r.rooms_id
        WHERE t.users_id = $1
          AND ($2 = '' OR t.created_at::date >= $2::date)
          AND ($3 = '' OR t.created_at::date <= $3::date)
          AND ($4 = '' OR t.status = $4::tx_status_enum)
          AND ($5 = '' OR r.type = $5::room_type)
    	`
		if err := db.QueryRow(countQuery, userID, startDateStr, endDateStr, status, typeStr).Scan(&totalData); err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Message: "Failed to count transactions: " + err.Error()})
		}
	case "admin":
		countQuery := `
        SELECT COUNT(DISTINCT t.tx_id)
        FROM transactions t
        JOIN detail_transaction dt ON t.tx_id = dt.tx_id
        JOIN rooms r ON dt.rooms_id = r.rooms_id
          AND ($1 = '' OR t.created_at::date >= $1::date)
          AND ($2 = '' OR t.created_at::date <= $2::date)
          AND ($3 = '' OR t.status = $3::tx_status_enum)
          AND ($4 = '' OR r.type = $4::room_type)
    	`
		if err := db.QueryRow(countQuery, startDateStr, endDateStr, status, typeStr).Scan(&totalData); err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Message: "Failed to count transactions: " + err.Error()})
		}
	}

	totalPage := (totalData + pageSize - 1) / pageSize

	// validasi role
	var rows *sql.Rows
	switch role {
	case "user":
		// Query data dengan filter + pagination
		rows, err = db.Query(`
        SELECT 
            t.tx_id, t.name, t.no_hp, t.company, t.total, t.status, t.created_at, t.updated_at,
            r.rooms_id, dt.sub_total_snacks, dt.sub_total_price_room, dt.snacks_id,
            r.name, r.type, r.price_perhour
        FROM transactions t
        JOIN detail_transaction dt ON t.tx_id = dt.tx_id
        JOIN rooms r ON dt.rooms_id = r.rooms_id
        WHERE t.users_id = $1
          AND ($2 = '' OR t.created_at::date >= $2::date)
          AND ($3 = '' OR t.created_at::date <= $3::date)
          AND ($4 = '' OR t.status = $4::tx_status_enum)
          AND ($5 = '' OR r.type = $5::room_type)
        ORDER BY t.created_at DESC
        LIMIT $6 OFFSET $7
    	`, userID, startDateStr, endDateStr, status, typeStr, pageSize, offset)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Message: "Failed to get transactions data: " + err.Error()})
		}
	case "admin":
		// Query data dengan filter + pagination
		rows, err = db.Query(`
        SELECT 
            t.tx_id, t.name, t.no_hp, t.company, t.total, t.status, t.created_at, t.updated_at,
            r.rooms_id, dt.sub_total_snacks, dt.sub_total_price_room, dt.snacks_id,
            r.name, r.type, r.price_perhour
        FROM transactions t
        JOIN detail_transaction dt ON t.tx_id = dt.tx_id
        JOIN rooms r ON dt.rooms_id = r.rooms_id
          AND ($1 = '' OR t.created_at::date >= $1::date)
          AND ($2 = '' OR t.created_at::date <= $2::date)
          AND ($3 = '' OR t.status = $3::tx_status_enum)
          AND ($4 = '' OR r.type = $4::room_type)
        ORDER BY t.created_at DESC
        LIMIT $5 OFFSET $6
    	`, startDateStr, endDateStr, status, typeStr, pageSize, offset)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Message: "Failed to get transactions data: " + err.Error()})
		}
	}
	defer rows.Close()

	// Map transaksi untuk menggabungkan multiple rooms
	transactionMap := make(map[uuid.UUID]*models.TransactionResponse)
	orderedKeys := []uuid.UUID{} // Untuk menyimpan urutan transaksi

	for rows.Next() {
		var (
			txID          uuid.UUID
			name          string
			phone         string
			company       string
			total         float64
			statusVal     string
			createdAt     string
			updatedAt     string
			roomID        int
			subTotalSnack float64
			subTotalRoom  float64
			snackID       any
			roomName      string
			roomType      string
			pricePerHour  float64
		)

		if err := rows.Scan(
			&txID, &name, &phone, &company, &total, &statusVal, &createdAt, &updatedAt,
			&roomID, &subTotalSnack, &subTotalRoom, &snackID,
			&roomName, &roomType, &pricePerHour,
		); err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Message: "Failed to scan data: " + err.Error()})
		}

		// Jika transaksi belum ada di map, buat entry baru
		if _, exists := transactionMap[txID]; !exists {
			transactionMap[txID] = &models.TransactionResponse{
				ID:          txID,
				Name:        name,
				PhoneNumber: phone,
				Company:     company,
				Total:       total,
				Status:      statusVal,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
				Rooms:       []models.Rooms{},
			}
			orderedKeys = append(orderedKeys, txID) // simpan urutan
		}

		// Tambahkan room ke transaksi terkait
		transactionMap[txID].Rooms = append(transactionMap[txID].Rooms, models.Rooms{
			ID:            roomID,
			SubTotalSnack: subTotalSnack,
			SubTotalRoom:  subTotalRoom,
			Snack:         models.Snack{ID: snackID},
			Name:          roomName,
			Type:          roomType,
			PricePerHour:  pricePerHour,
		})
	}

	// Bangun slice transactions sesuai urutan
	transactions := []models.TransactionResponse{}
	for _, id := range orderedKeys {
		transactions = append(transactions, *transactionMap[id])
	}

	return c.JSON(http.StatusOK, utils.ListResponse{
		Message:   "Success",
		Data:      transactions,
		Page:      page,
		PageSize:  pageSize,
		TotalData: totalData,
		TotalPage: totalPage,
	})
}

// ----------------------- UNTUK GET 1 TRANSAKSI -> 1 ROOM SAJA (HARDCODE ROOM[0]) -----------------------
// func HistoryReservation(c echo.Context, db *sql.DB) error {
// 	// ambil query parameter
// 	startDateStr := c.QueryParam("startDate")
// 	endDateStr := c.QueryParam("endDate")
// 	status := c.QueryParam("status")
// 	typeStr := c.QueryParam("type")
// 	pageStr := c.QueryParam("page")
// 	pageSizeStr := c.QueryParam("pageSize")

// 	var page, pageSize int
// 	var err error
// 	if pageStr != "" {
// 		if page, err = strconv.Atoi(pageStr); err != nil {
// 			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
// 				Message: "Invalid page parameter",
// 			})
// 		}
// 	} else {
// 		page = 1 // Default to page 1 if not provided
// 	}
// 	if pageSizeStr != "" {
// 		if pageSize, err = strconv.Atoi(pageSizeStr); err != nil {
// 			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
// 				Message: "Invalid pageSize parameter",
// 			})
// 		}
// 	} else {
// 		pageSize = 10 // Default to 10 items per page if not provided
// 	}
// 	offset := (page - 1) * pageSize

// 	// ambil claim token dari context
// 	claims := c.Get("client").(jwt.MapClaims)
// 	// ambil user_id dari klaim token
// 	userIDfloat, ok := claims["id"].(float64)
// 	if !ok {
// 		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
// 			Message: "Invalid token claims",
// 		})
// 	}
// 	userID := int(userIDfloat) // konversi ke int

// 	// Hitung total data
// 	var totalData int
// 	countQuery := `
//         SELECT COUNT(DISTINCT t.tx_id)
//         FROM transactions t
//         JOIN detail_transaction dt ON t.tx_id = dt.tx_id
//         JOIN rooms r ON dt.rooms_id = r.rooms_id
//         WHERE t.users_id = $1
//           AND ($2 = '' OR t.created_at::date >= $2::date)
//           AND ($3 = '' OR t.created_at::date <= $3::date)
//           AND ($4 = '' OR t.status = $4::tx_status_enum)
//           AND ($5 = '' OR r.type = $5::room_type)
//     `
// 	if err := db.QueryRow(countQuery, userID, startDateStr, endDateStr, status, typeStr).Scan(&totalData); err != nil {
// 		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Message: "Failed to count transactions: " + err.Error()})
// 	}

// 	// Hitung total halaman
// 	totalPages := (totalData + pageSize - 1) / pageSize

// 	// Query data
// 	rows, err := db.Query(`
//         SELECT
//             t.tx_id, t.name, t.no_hp, t.company, t.total, t.status, t.created_at, t.updated_at,
//             r.rooms_id, dt.sub_total_snacks, dt.sub_total_price_room, dt.snacks_id,
//             r.name, r.type, r.price_perhour
//         FROM transactions t
//         JOIN detail_transaction dt ON t.tx_id = dt.tx_id
//         JOIN rooms r ON dt.rooms_id = r.rooms_id
//         WHERE t.users_id = $1
//           AND ($2 = '' OR t.created_at::date >= $2::date)
//           AND ($3 = '' OR t.created_at::date <= $3::date)
//           AND ($4 = '' OR t.status = $4::tx_status_enum)
//           AND ($5 = '' OR r.type = $5::room_type)
//         ORDER BY t.created_at DESC
//         LIMIT $6 OFFSET $7
//     `, userID, startDateStr, endDateStr, status, typeStr, pageSize, offset)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Message: "Failed to get transactions data: " + err.Error()})
// 	}
// 	defer rows.Close()

// 	// buat slice untuk menyimpan data
// 	var transactions []models.TransactionResponse
// 	// iterasi melalui hasil query dan masukkan ke dalam slice
// 	for rows.Next() {
// 		var transaction models.TransactionResponse
// 		// inisialisasi minimal 1 elemen di slice Rooms
// 		transaction.Rooms = make([]models.Rooms, 1)
// 		if err := rows.Scan(
// 			&transaction.ID,
// 			&transaction.Name,
// 			&transaction.PhoneNumber,
// 			&transaction.Company,
// 			&transaction.Total,
// 			&transaction.Status,
// 			&transaction.CreatedAt,
// 			&transaction.UpdatedAt,
// 			&transaction.Rooms[0].ID,
// 			&transaction.Rooms[0].SubTotalSnack,
// 			&transaction.Rooms[0].SubTotalRoom,
// 			&transaction.Rooms[0].Snack.ID,
// 			&transaction.Rooms[0].Name,
// 			&transaction.Rooms[0].Type,
// 			&transaction.Rooms[0].PricePerHour,
// 		); err != nil {
// 			if err == sql.ErrNoRows {
// 				return c.JSON(http.StatusNotFound, utils.ErrorResponse{
// 					Message: "Transactions not found",
// 				})
// 			}
// 			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
// 				Message: "Failed to get transactions data: " + err.Error(),
// 			})
// 		}
// 		transactions = append(transactions, transaction)
// 	}

// 	return c.JSON(http.StatusOK, utils.ListResponse{
// 		Message:   "Success",
// 		Data:      transactions,
// 		Page:      page,
// 		PageSize:  pageSize,
// 		TotalPage: totalPages,
// 		TotalData: totalData,
// 	})
// }

// @Summary Get reservation by ID
// @Description Get reservation by ID
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Reservation ID"
// @Success 200 {object} utils.SuccessResponse{data=models.CalculationResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /reservations/{id} [get]
func GetReservationByID(c echo.Context, db *sql.DB) error {
	// ambil path parameter
	id := c.Param("id")

	// query data
	row := db.QueryRow(`
	SELECT t.tx_id, t.name, t.no_hp, t.company, t.total, t.status, t.created_at, t.updated_at,
		r.rooms_id, dt.sub_total_snacks, dt.sub_total_price_room, dt.snacks_id, dt.start_time, dt.end_time, dt.participants, dt.price_snack_perpack,
		r.name, r.type, r.price_perhour, r.capacity, r.img_path,
		s.name AS snack_name, 
		s.category AS snack_category
	FROM transactions t
	JOIN detail_transaction dt ON t.tx_id = dt.tx_id
	JOIN rooms r ON dt.rooms_id = r.rooms_id
	LEFT JOIN snacks s on dt.snacks_id = s.snacks_id
	WHERE t.tx_id = $1
`, id)
	var (
		txID          uuid.UUID
		name          string
		phone         string
		company       string
		total         float64
		statusVal     string
		createdAt     string
		updatedAt     string
		roomID        int
		subTotalSnack float64
		subTotalRoom  float64
		snackID       any
		startTime     string
		endTime       string
		participants  int
		priceSnack    float64
		roomName      string
		roomType      string
		pricePerHour  float64
		imgPath       string
		capacity      int
		snackName     sql.NullString
		snackCategory sql.NullString
	)
	if err := row.Scan(
		&txID,
		&name,
		&phone,
		&company,
		&total,
		&statusVal,
		&createdAt,
		&updatedAt,
		&roomID,
		&subTotalSnack,
		&subTotalRoom,
		&snackID,
		&startTime,
		&endTime,
		&participants,
		&priceSnack,
		&roomName,
		&roomType,
		&pricePerHour,
		&capacity,
		&imgPath,
		&snackName,
		&snackCategory,
	); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Message: "Transactions not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to get transactions data: " + err.Error(),
		})
	}

	// handle kondisi snackID
	if snackID == nil {
		snackID = int64(0)
	}

	// konversi startTime dan endTime ke layout yang diinginkan
	// Parse dari RFC3339 bawaan PostgreSQL
	parsedTime, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid start time format (RFC3339 parse failed): " + err.Error(),
		})
	}

	// Format ulang ke layout yang diminta fungsi konversi
	startTime = parsedTime.Format("2006-01-02 15:04:05.000 -0700")

	parsedTime, err = time.Parse(time.RFC3339, endTime)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid end time format (RFC3339 parse failed): " + err.Error(),
		})
	}
	endTime = parsedTime.Format("2006-01-02 15:04:05.000 -0700")

	// calculate duration
	startTimeTime, err := utils.StringToTimestamptz(startTime)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid start time format: " + err.Error(),
		})
	}
	endTimeTime, err := utils.StringToTimestamptz(endTime)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid end time format: " + err.Error(),
		})
	}
	duration := utils.CalculateDuration(startTimeTime, endTimeTime)

	// pisahkan data sesuai struct CalculationResponse
	var data models.CalculationResponse
	data.Rooms = append(data.Rooms, models.RoomCalculation{
		Name:          roomName,
		Type:          roomType,
		PricePerHour:  pricePerHour,
		SubTotalSnack: subTotalSnack,
		SubTotalRoom:  subTotalRoom,
		Capacity:      capacity,
		ImgPath:       imgPath,
		Snacks: models.SnacksCalculation{
			ID:       int(snackID.(int64)),
			Category: snackCategory.String,
			Name:     snackName.String,
			Price:    priceSnack,
		},
	})
	data.PersonalData.Name = name
	data.PersonalData.NoHp = phone
	data.PersonalData.Company = company
	data.PersonalData.StartTime = startTime
	data.PersonalData.EndTime = endTime
	data.PersonalData.Duration = int(duration.Hours())
	data.PersonalData.Participants = participants
	data.Total = total

	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Success",
		Data:    data,
	})
}

// @Summary UpdateReservationStatus updates the status of a reservation
// @Description Update the status of a reservation with the provided ID
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Reservation ID"
// @Param status body models.StatusReservation true "Status to update"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /reservations/status/{id} [patch]
func UpdateReservationStatus(c echo.Context, db *sql.DB) error {
	// ambil status dari payload
	id := c.Param("id")
	var status models.StatusReservation
	if err := c.Bind(&status); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request body",
		})
	}

	tx, err := db.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to start transaction: " + err.Error(),
		})
	}
	defer tx.Rollback()

	// cek apakah t.created_at dan t.updated_at bernilai sama, jika beda maka status tidak boleh diubah
	row := tx.QueryRow(`SELECT created_at, updated_at FROM transactions t WHERE t.tx_id = $1`, id)
	var createdAt, updatedAt time.Time
	if err := row.Scan(&createdAt, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Message: "Transactions not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to get transactions data: " + err.Error(),
		})
	}
	if createdAt != updatedAt {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Status cannot be updated because the transaction has been processed",
		})
	}

	// update status
	_, err = tx.Exec(`UPDATE transactions SET status = $1 WHERE tx_id = $2`, status.Status, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to update transaction status: " + err.Error(),
		})
	}

	if err := tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to commit transaction: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Success",
	})

}

// API handler: push request ke antrean Redis
// @Summary CreateReservation creates a new reservation
// @Description Create a new reservation with the provided details
// @Tags reservations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ReservationRequest true "Reservation details"
// @Success 202 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /reservations [post]
func CreateReservation(c echo.Context, rdb *redis.Client) error {
	// ambil claim token
	claims := c.Get("client").(jwt.MapClaims)
	userIDfloat, ok := claims["id"].(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid token claims",
		})
	}
	userIDInt := int(userIDfloat)

	// ambil body request
	var request models.ReservationRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request body",
		})
	}

	if request.UserID != userIDInt {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid user ID",
		})
	}

	// Serialize request
	// menyiapkan data agar bisa di push ke antrean
	data, err := json.Marshal(request) // marshal itu untuk mengubah struct menjadi json yang bisa dibaca oleh redis
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to serialize request: " + err.Error(),
		})
	}

	// Push ke antrean Redis
	err = rdb.RPush(c.Request().Context(), "reservation:queue", data).Err()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to push to queue: " + err.Error(),
		})
	}

	return c.JSON(http.StatusAccepted, utils.SuccessResponse{
		Message: "Reservation request queued",
	})
}

// Worker processor: eksekusi logic reservasi
func ProcessReservation(db *sql.DB, request models.ReservationRequest) error {
	// mulai transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	// sebelum ambil data,
	// cek apakah startTime dan endTime beririsan dengan startTimeDB dan endTimeDB
	startTime, err := utils.StringToTimestamptz(request.Rooms[0].StartTime)
	if err != nil {
		return fmt.Errorf("invalid start time format")
	}
	endTime, err := utils.StringToTimestamptz(request.Rooms[0].EndTime)
	if err != nil {
		return fmt.Errorf("invalid end time format")
	}

	var overlaps bool
	err = tx.QueryRow(`
    SELECT EXISTS (
        SELECT 1
        FROM detail_transaction dt
        JOIN transactions t ON dt.tx_id = t.tx_id
        WHERE rooms_id = $1
          AND t.status != 'canceled'
          AND (dt.start_time, dt.end_time) OVERLAPS ($2, $3)
    )
	`, request.Rooms[0].ID, startTime, endTime).Scan(&overlaps)

	if err != nil {
		return fmt.Errorf("check overlap: %w", err)
	}

	if overlaps {
		return fmt.Errorf("room is already reserved")
	}

	// cek apakah startTime lebih besar atau sama dengan endTime
	if !startTime.Before(endTime) {
		return fmt.Errorf("start time must be before end time")
	}

	// ambil data snack
	// snackInt := int(request.Rooms[0].SnackID.(float64))
	var snack models.Snacks
	if *request.Rooms[0].SnackID != 0 {
		row := tx.QueryRow("SELECT snacks_id, name, price, category FROM snacks WHERE snacks_id = $1", request.Rooms[0].SnackID)
		if err := row.Scan(&snack.ID, &snack.Name, &snack.Price, &snack.Category); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("snack not found")
			}
			return fmt.Errorf("get snack data: %w", err)
		}
	} else if *request.Rooms[0].SnackID == 0 {
		request.Rooms[0].SnackID = nil // set nil karena di db nullable
	}

	// ambil data room
	var room models.CURoomRequest
	row := tx.QueryRow("SELECT name, type, price_perhour, capacity, img_path FROM rooms WHERE rooms_id = $1", request.Rooms[0].ID)
	if err := row.Scan(&room.Name, &room.Type, &room.PricePerHour, &room.Capacity, &room.ImgPath); err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("room not found")
		}
		return fmt.Errorf("get room data: %w", err)
	}

	// hitung durasi & total
	duration := utils.CalculateDuration(startTime, endTime)

	snackPrice, _ := (snack.Price).Float64()
	subTotalSnack := snackPrice * float64(request.Rooms[0].Participants)
	subTotalRoom := room.PricePerHour * float64(duration.Hours())
	total := subTotalSnack + subTotalRoom

	// generate UUID
	uuidTx := uuid.New().String()
	uuidDetailTx := uuid.New().String()

	// insert transactions
	_, err = tx.Exec(`INSERT INTO transactions (tx_id, users_id, name, no_hp, company, note, total) 
					  VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		uuidTx, request.UserID, request.Name, request.PhoneNumber, request.Company, request.Notes, total)
	if err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}

	// insert detail_transaction
	_, err = tx.Exec(`INSERT INTO detail_transaction 
		(detail_tx_id, tx_id, rooms_id, start_time, end_time, participants, snacks_id, 
		sub_total_snacks, sub_total_price_room, price_snack_perpack, price_room_perhour) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		uuidDetailTx, uuidTx, request.Rooms[0].ID, startTime, endTime,
		request.Rooms[0].Participants, request.Rooms[0].SnackID,
		subTotalSnack, subTotalRoom, snack.Price, room.PricePerHour)
	if err != nil {
		return fmt.Errorf("insert detail transaction: %w", err)
	}

	// commit
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
