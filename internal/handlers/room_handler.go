package handlers

import (
	"database/sql"
	"e_meeting/config"
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
)

func InitRoomHandler(e *echo.Group, dbConn *sql.DB) {
	e.GET("/rooms", func(c echo.Context) error {
		return GetRooms(c, dbConn)
	})
	e.POST("/rooms", func(c echo.Context) error {
		return CreateRoom(c, dbConn)
	})
	e.PUT("/rooms/:id", func(c echo.Context) error {
		return UpdateRoom(c, dbConn)
	})
	e.DELETE("/rooms/:id", func(c echo.Context) error {
		return DeleteRoom(c, dbConn)
	})
	e.GET("/rooms/:id/reservations/:date", func(c echo.Context) error {
		return GetRoomSchedule(c, dbConn)
	})
	e.GET("/rooms/reservations", func(c echo.Context) error {
		return GetRoomScheduleAdmin(c, dbConn)
	})
}

// Summary GetRooms retrieves a list of rooms with optional filters
// Description Get a list of rooms with optional filters for name, type, and capacity
// Tags rooms
// Accept json
// Produce json
// Security BearerAuth
// Param name query string false "Room name"
// Param type query string false "Room type"
// Param capacity query string false "Room capacity"
// Param page query int false "Page number"
// Param pageSize query int false "Number of items per page"
// Success 200 {object} utils.ListResponse{data=[]models.Room}
// Failure 400 {object} utils.ErrorResponse
// Failure 500 {object} utils.ErrorResponse
// Router /rooms [get]
func GetRooms(c echo.Context, db *sql.DB) error {
	// ambil query parameter untuk filter nama ruangan
	roomName := c.QueryParam("name")
	roomType := c.QueryParam("type")
	capacityStr := c.QueryParam("capacity")
	var capacity int
	var err error
	if capacityStr != "" {
		capacity, err = strconv.Atoi(capacityStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Message: "Invalid capacity parameter",
			})
		}
	}
	pageStr := c.QueryParam("page")
	pageSizeStr := c.QueryParam("pageSize")
	var page, pageSize int
	if pageStr != "" {
		if page, err = strconv.Atoi(pageStr); err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Message: "Invalid page parameter",
			})
		}
	} else {
		page = 1 // Default to page 1 if not provided
	}
	if pageSizeStr != "" {
		if pageSize, err = strconv.Atoi(pageSizeStr); err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Message: "Invalid pageSize parameter",
			})
		}
	} else {
		pageSize = 10 // Default to 10 items per page if not provided
	}
	offset := (page - 1) * pageSize

	// ambil daftar ruangan dengan filter
	rows, err := db.Query(`
    	SELECT rooms_id, name, type, price_perhour, capacity, img_path, created_at, updated_at 
    	FROM rooms
    	WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
    		AND ($2 = '' OR type = $2::room_type)
      		AND ($3 = 0 OR capacity >= $3)
    	ORDER BY rooms_id
    	LIMIT $4 OFFSET $5
		`, roomName, roomType, capacity, pageSize, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to retrieve rooms: " + err.Error(),
		})
	}
	defer rows.Close()

	var rooms []models.RoomResponse
	for rows.Next() {
		var room models.RoomResponse
		if err := rows.Scan(&room.ID, &room.Name, &room.Type, &room.PricePerHour, &room.Capacity, &room.ImgPath, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Failed to retrieve rooms: " + err.Error(),
			})
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to retrieve rooms: " + err.Error(),
		})
	}
	// jika tidak ada ruangan yang ditemukan, kembalikan response dengan message
	if len(rooms) == 0 {
		return c.JSON(http.StatusOK, utils.SuccessResponse{
			Message: "No rooms found",
			Data:    []models.Room{},
		})
	}
	// hitung total data untuk pagination
	var totalData int
	err = db.QueryRow(`
    	SELECT COUNT(*) 
    	FROM rooms
    	WHERE ($1 = '' OR name ILIKE '%' || $1 || '%')
      		AND ($2 = '' OR type = $2::room_type)
      		AND ($3 = 0 OR capacity >= $3)
		`, roomName, roomType, capacity).Scan(&totalData)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to count total rooms: " + err.Error(),
		})
	}
	// hitung total halaman
	totalPage := (totalData + pageSize - 1) / pageSize

	// jika ada ruangan yang ditemukan, kembalikan response dengan daftar ruangan
	return c.JSON(http.StatusOK, utils.ListResponse{
		Message:   "List of rooms retrieved successfully",
		Data:      rooms,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
		TotalData: totalData,
	})
}

// @Summary CreateRoom creates a new room
// @Description Create a new room with the provided details
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CURoomRequest true "Room details"
// @Success 201 {object} utils.SuccessResponse{data=nil}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /rooms [post]
func CreateRoom(c echo.Context, db *sql.DB) error {
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
	// ambil status dari klaim token
	status, ok := claims["status"].(string)
	// jika status bukan active, kembalikan response unauthorized
	if !ok || status != "active" {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Unauthorized access",
		})
	}

	// ambil data dari request body
	var room models.CURoomRequest
	if err := c.Bind(&room); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request body: " + err.Error(),
		})
	}
	// validasi data
	if room.Name == "" || room.Type == "" || room.PricePerHour <= 0 || room.Capacity <= 0 {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Name, type, price per hour, and capacity are required and must be valid",
		})
	}

	imgUrl := room.ImgPath
	if imgUrl != "" {
		domain := config.New().Domain

		// validasi url
		parsedURL, err := url.Parse(room.ImgPath)
		if err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Message: "Invalid URL: " + err.Error(),
			})
		}

		// ambil baseDomain dari request
		baseDomain := parsedURL.Scheme + "://" + parsedURL.Host

		// validasi baseDomain
		if baseDomain != domain {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Message: "Invalid base domain: " + baseDomain,
			})
		}

		// pindahkan file dari temp ke uploads
		data, err := UploadFile(c, room.ImgPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Failed to upload file: " + err.Error(),
			})
		}

		// ambil data dari channel
		fmt.Println("fileRequest: ", data)
		imgUrl = data.ImageURL
	}

	// simpan data room ke database
	_, err := db.Exec(`INSERT INTO rooms (name, type, price_perhour, capacity, img_path) VALUES ($1, $2, $3, $4, $5)`,
		room.Name, room.Type, room.PricePerHour, room.Capacity, imgUrl)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to create room: " + err.Error(),
		})
	}
	// kembalikan response dengan message
	return c.JSON(http.StatusCreated, utils.SuccessResponse{
		Message: "Room created successfully",
		Data:    nil,
	})
}

// @Summary UpdateRoom updates an existing room
// @Description Update an existing room with the provided details
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Room ID"
// @Param request body models.CURoomRequest true "Room details"
// @Success 200 {object} utils.SuccessResponse{data=nil}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /rooms/{id} [put]
func UpdateRoom(c echo.Context, db *sql.DB) error {
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
	// ambil status dari klaim token
	status, ok := claims["status"].(string)
	// jika status bukan active, kembalikan response unauthorized
	if !ok || status != "active" {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Unauthorized access",
		})
	}

	// ambil id dari parameter
	id := c.Param("id")
	// konversi id ke int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid room ID: " + err.Error(),
		})
	}

	// ambil data dari request body
	var room models.CURoomRequest
	if err := c.Bind(&room); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request body: " + err.Error(),
		})
	}
	// validasi data
	if room.Name == "" || room.Type == "" || room.PricePerHour <= 0 || room.Capacity <= 0 {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Name, type, price per hour, and capacity are required and must be valid",
		})
	}

	// ambil data room image_path dari database
	var dbImgPath string
	err = db.QueryRow(`SELECT img_path FROM rooms WHERE rooms_id = $1`, idInt).Scan(&dbImgPath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to get room image path: " + err.Error(),
		})
	}

	// jika room image_path tidak sama dengan request image_path, upload file
	imgUrl := dbImgPath
	if dbImgPath != room.ImgPath {
		domain := config.New().Domain

		// validasi url
		parsedURL, err := url.Parse(room.ImgPath)
		if err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Message: "Invalid URL: " + err.Error(),
			})
		}

		// ambil baseDomain dari request
		baseDomain := parsedURL.Scheme + "://" + parsedURL.Host

		// validasi baseDomain
		if baseDomain != domain {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Message: "Invalid base domain: " + baseDomain,
			})
		}

		// pindahkan file dari temp ke uploads
		data, err := UploadFile(c, room.ImgPath)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Failed to upload file: " + err.Error(),
			})
		}

		// ambil data dari channel
		fmt.Println("fileRequest: ", data)
		imgUrl = data.ImageURL
	}

	// simpan data room ke database
	_, err = db.Exec(`UPDATE rooms SET name = $1, type = $2, price_perhour = $3, capacity = $4, img_path = $5 WHERE rooms_id = $6`,
		room.Name, room.Type, room.PricePerHour, room.Capacity, imgUrl, idInt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to update room: " + err.Error(),
		})
	}
	// kembalikan response dengan message
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Room updated successfully",
		Data:    nil,
	})
}

// @Summary DeleteRoom deletes a room
// @Description Delete a room with the provided ID
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Room ID"
// @Success 200 {object} utils.SuccessResponse{data=nil}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /rooms/{id} [delete]
func DeleteRoom(c echo.Context, db *sql.DB) error {
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
	// ambil status dari klaim token
	status, ok := claims["status"].(string)
	// jika status bukan active, kembalikan response unauthorized
	if !ok || status != "active" {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Unauthorized access",
		})
	}

	// ambil id dari parameter
	id := c.Param("id")
	// konversi id ke int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid room ID: " + err.Error(),
		})
	}

	// cek apakah room dengan id tersebut sudah punya transaksi
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM detail_transaction WHERE rooms_id = $1`, idInt).Scan(&count)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to check room transactions: " + err.Error(),
		})
	}
	if count > 0 {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Cannot delete room with transactions",
		})
	}

	// hapus data room dari database dan jika no row affected, kembalikan response not found
	result, err := db.Exec(`DELETE FROM rooms WHERE rooms_id = $1`, idInt)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to delete room: " + err.Error(),
		})
	}
	res, _ := result.RowsAffected()
	if res == 0 {
		return c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Message: "Room not found",
		})
	}

	// kembalikan response dengan message
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Room deleted successfully",
		Data:    nil,
	})
}

// @Summary GetRoomSchedule gets the schedule of a room
// @Description Get the schedule of a room with the provided ID
// @Tags rooms schedule
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Room ID"
// @Param date path string true "Date"
// @Success 200 {object} utils.SuccessResponse{data=models.RoomScheduleResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /rooms/{id}/reservations/{date} [get]
func GetRoomSchedule(c echo.Context, db *sql.DB) error {
	// ambil id dari parameter
	id := c.Param("id")
	dateFilter := c.Param("date")
	// konversi id ke int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid room ID: " + err.Error(),
		})
	}

	// ambil data room dari database yang difilter berdasarkan id dan status
	rows, err := db.Query(`select t.status, dt.start_time, dt.end_time 
			from transactions t
			join detail_transaction dt on t.tx_id = dt.tx_id
			where dt.rooms_id = $1 and t.status != 'canceled' AND dt.start_time >= $2::date AND dt.start_time < ($2::date + INTERVAL '1 day')`,
		idInt, dateFilter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to retrieve room schedule: " + err.Error(),
		})
	}
	defer rows.Close()
	var rooms []models.RoomSchedule
	for rows.Next() {
		var room models.RoomSchedule
		if err := rows.Scan(&room.Status, &room.StartTime, &room.EndTime); err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Failed to scan room schedule: " + err.Error(),
			})
		}
		rooms = append(rooms, room)
	}

	// ambil data nama room
	var roomName string
	err = db.QueryRow(`select name from rooms where rooms_id = $1`, idInt).Scan(&roomName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to retrieve room name: " + err.Error(),
		})
	}

	// count rows yang berhasil di dapat
	count := len(rooms)
	schedule := models.RoomScheduleResponse{
		RoomName:    roomName,
		Schedule:    rooms,
		TotalBooked: count,
	}
	// kembalikan response dengan data room
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Room schedule retrieved successfully",
		Data:    schedule,
	})
}

// @Summary GetRoomScheduleAdmin gets the schedule of a room
// @Description Get the schedule of a room
// @Tags rooms schedule
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_time query string false "Start time"
// @Param end_time query string false "End time"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {object} utils.ListResponse{data=[]models.RoomScheduleAdminResponse}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /rooms/reservations [get]
func GetRoomScheduleAdmin(c echo.Context, db *sql.DB) error {
	// Ambil role dari klaim token
	claims := c.Get("client").(jwt.MapClaims)
	role, ok := claims["role"].(string)
	if !ok || role != "admin" {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Unauthorized access",
		})
	}

	// Ambil query parameter
	startTime := c.QueryParam("start_time")
	endTime := c.QueryParam("end_time")
	pageStr := c.QueryParam("page")
	pageSizeStr := c.QueryParam("pageSize")

	var page, pageSize int
	var err error
	if pageStr != "" {
		if page, err = strconv.Atoi(pageStr); err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Message: "Invalid page parameter",
			})
		}
	} else {
		page = 1
	}
	if pageSizeStr != "" {
		if pageSize, err = strconv.Atoi(pageSizeStr); err != nil {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Message: "Invalid pageSize parameter",
			})
		}
	} else {
		pageSize = 5
	}
	offset := (page - 1) * pageSize

	// Hitung total room sesuai filter
	var totalData int
	countQuery := `
		SELECT COUNT(DISTINCT r.rooms_id)
		FROM detail_transaction dt
		JOIN transactions t ON dt.tx_id = t.tx_id
		JOIN rooms r ON dt.rooms_id = r.rooms_id
		WHERE t.status != 'canceled'
		AND ($1 = '' OR dt.start_time::date >= $1::date)
		AND ($2 = '' OR dt.end_time::date <= $2::date)
	`
	if err := db.QueryRow(countQuery, startTime, endTime).Scan(&totalData); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to count room schedule: " + err.Error(),
		})
	}

	// Ambil daftar room sesuai pagination
	roomQuery := `
		SELECT DISTINCT r.rooms_id, r.name
		FROM detail_transaction dt
		JOIN transactions t ON dt.tx_id = t.tx_id
		JOIN rooms r ON dt.rooms_id = r.rooms_id
		WHERE t.status != 'canceled'
		AND ($1 = '' OR dt.start_time::date >= $1::date)
		AND ($2 = '' OR dt.end_time::date <= $2::date)
		ORDER BY r.name
		LIMIT $3 OFFSET $4
	`
	roomRows, err := db.Query(roomQuery, startTime, endTime, pageSize, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to retrieve room list: " + err.Error(),
		})
	}
	defer roomRows.Close()

	var roomIDs []int
	roomMap := make(map[int]models.RoomScheduleAdminResponse)
	for roomRows.Next() {
		var id int
		var name string
		if err := roomRows.Scan(&id, &name); err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Failed to scan room list: " + err.Error(),
			})
		}
		roomIDs = append(roomIDs, id)
		roomMap[id] = models.RoomScheduleAdminResponse{
			RoomName: name,
			Schedule: []models.RoomScheduleAdmin{},
		}
	}

	// Ambil semua schedule untuk room yang ada di halaman ini
	if len(roomIDs) > 0 {
		scheduleQuery := `
			SELECT r.rooms_id, t.company, dt.start_time, dt.end_time, COALESCE(t.status::text, '') AS status
			FROM detail_transaction dt
			JOIN transactions t ON dt.tx_id = t.tx_id
			JOIN rooms r ON dt.rooms_id = r.rooms_id
			WHERE t.status != 'canceled'
			AND r.rooms_id = ANY($1)
			AND ($2 = '' OR dt.start_time::date >= $2::date)
			AND ($3 = '' OR dt.end_time::date <= $3::date)
			ORDER BY dt.start_time ASC
		`
		schedRows, err := db.Query(scheduleQuery, pq.Array(roomIDs), startTime, endTime)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Failed to retrieve schedules: " + err.Error(),
			})
		}
		defer schedRows.Close()

		for schedRows.Next() {
			var roomID int
			var startTime time.Time
			var endTime time.Time
			var sch models.RoomScheduleAdmin
			if err := schedRows.Scan(&roomID, &sch.Company, &startTime, &endTime, &sch.Status); err != nil {
				return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
					Message: "Failed to scan schedule: " + err.Error(),
				})
			}

			// Format time.Time ke string sesuai format JSON
			sch.StartTime = startTime.Format("2006-01-02 15:04:05")
			sch.EndTime = endTime.Format("2006-01-02 15:04:05")

			// Hitung StatusProgress
			now := time.Now()
			if now.After(endTime) {
				sch.StatusProgress = "Done"
			} else if now.After(startTime) && now.Before(endTime) {
				sch.StatusProgress = "In Progress"
			} else {
				sch.StatusProgress = "Up Coming"
			}
			room := roomMap[roomID]
			room.Schedule = append(room.Schedule, sch)
			roomMap[roomID] = room
		}
	}

	// Susun hasil akhir
	var reservation []models.RoomScheduleAdminResponse
	for _, v := range roomMap {
		reservation = append(reservation, v)
	}

	totalPage := (totalData + pageSize - 1) / pageSize

	return c.JSON(http.StatusOK, utils.ListResponse{
		Message:   "Room schedule retrieved successfully",
		Data:      reservation,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
		TotalData: totalData,
	})
}

// func GetRoomScheduleAdmin(c echo.Context, db *sql.DB) error {
// 	// ambil role dari klaim token
// 	claims := c.Get("client").(jwt.MapClaims)
// 	role, ok := claims["role"].(string)
// 	// jika role bukan admin, kembalikan response unauthorized
// 	if !ok || role != "admin" {
// 		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
// 			Message: "Unauthorized access",
// 		})
// 	}

// 	// ambil query parameter
// 	startTime := c.QueryParam("start_time")
// 	endTime := c.QueryParam("end_time")
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
// 		pageSize = 5 // Default to 5 items per page if not provided
// 	}
// 	offset := (page - 1) * pageSize

// 	// Hitung total data sesuai filter
// 	var totalData int
// 	countQuery := `
// 		SELECT COUNT(DISTINCT r.rooms_id)
// 		FROM detail_transaction dt
// 		JOIN transactions t ON dt.tx_id = t.tx_id
// 		JOIN rooms r ON dt.rooms_id = r.rooms_id
// 		WHERE t.status != 'canceled'
// 		AND ($1 = '' OR dt.start_time::date >= $1::date)
//         AND ($2 = '' OR dt.end_time::date <= $2::date)
// 	`
// 	if err := db.QueryRow(countQuery, startTime, endTime).Scan(&totalData); err != nil {
// 		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
// 			Message: "Failed to count room schedule: " + err.Error(),
// 		})
// 	}

// 	// Query data dengan filter & pagination
// 	query := `
// 		SELECT r.name, t.company, dt.start_time, dt.end_time
// 		FROM detail_transaction dt
// 		JOIN transactions t ON dt.tx_id = t.tx_id
// 		JOIN rooms r ON dt.rooms_id = r.rooms_id
// 		WHERE t.status != 'canceled'
// 		AND ($1 = '' OR dt.start_time::date >= $1::date)
//         AND ($2 = '' OR dt.end_time::date <= $2::date)
// 		ORDER BY dt.start_time ASC
// 		LIMIT $3 OFFSET $4
// 	`
// 	rows, err := db.Query(query, startTime, endTime, pageSize, offset)
// 	if err != nil {
// 		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
// 			Message: "Failed to retrieve room schedule: " + err.Error(),
// 		})
// 	}
// 	defer rows.Close()

// 	// map untuk group jadwal berdasarkan roomName
// 	roomMap := make(map[string][]models.RoomScheduleAdmin)

// 	for rows.Next() {
// 		var roomName string
// 		var item models.RoomScheduleAdmin
// 		if err := rows.Scan(&roomName, &item.Company, &item.StartTime, &item.EndTime); err != nil {
// 			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
// 				Message: "Failed to scan room schedule: " + err.Error(),
// 			})
// 		}
// 		roomMap[roomName] = append(roomMap[roomName], item)
// 	}

// 	var reservation []models.RoomScheduleAdminResponse
// 	for name, schedule := range roomMap {
// 		reservation = append(reservation, models.RoomScheduleAdminResponse{
// 			RoomName: name,
// 			Schedule: schedule,
// 		})
// 	}
// 	totalPage := (totalData + pageSize - 1) / pageSize

// 	// kembalikan response dengan data room
// 	return c.JSON(http.StatusOK, utils.ListResponse{
// 		Message:   "Room schedule retrieved successfully",
// 		Data:      reservation,
// 		Page:      page,
// 		PageSize:  pageSize,
// 		TotalPage: totalPage,
// 		TotalData: totalData,
// 	})
// }
