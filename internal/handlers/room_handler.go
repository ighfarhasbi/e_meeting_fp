package handlers

import (
	"database/sql"
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func InitRoomHandler(e *echo.Group, dbConn *sql.DB) {
	e.GET("/rooms", func(c echo.Context) error {
		return GetRooms(c, dbConn)
	})
}

// @Summary GetRooms retrieves a list of rooms with optional filters
// @Description Get a list of rooms with optional filters for name, type, and capacity
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param name query string false "Room name"
// @Param type query string false "Room type"
// @Param capacity query string false "Room capacity"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {object} utils.ListResponse{data=[]models.Room}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /rooms [get]
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

	var rooms []models.Room
	for rows.Next() {
		var room models.Room
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
