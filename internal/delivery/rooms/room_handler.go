package delivery

import (
	"e_meeting/config"
	"e_meeting/internal/entity"
	UploadFile "e_meeting/internal/handlers"
	"e_meeting/internal/models/request"
	repository "e_meeting/internal/repository/rooms"
	usecase "e_meeting/internal/usecase/rooms"
	"e_meeting/pkg/utils"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type RoomsHandler struct {
	uc *usecase.RoomsUsecase
}

func NewRoomHandler(e *echo.Group, uc *usecase.RoomsUsecase) {
	handler := &RoomsHandler{uc}
	e.GET("/rooms", handler.RoomsList)
	e.POST("/rooms", handler.CreateRoom)
	e.PUT("/rooms/:id", handler.UpdateRoom)
	e.DELETE("/rooms/:id", handler.DeleteRoom)
}

// @Summary Get Rooms List
// @Description Get a list of rooms with optional filters and pagination
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param roomName query string false "Filter by room name"
// @Param roomType query string false "Filter by room type"
// @Param capacity query int false "Filter by room capacity"
// @Param pageSize query int false "Number of items per page"
// @Param page query int false "Page number"
// @Success 200 {object} utils.SuccessResponse{data=[]entity.Rooms} "Rooms list retrieved successfully"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /rooms [get]
func (h *RoomsHandler) RoomsList(c echo.Context) error {
	// ambil query parameters
	roomName := c.QueryParam("roomName")
	roomType := c.QueryParam("roomType")
	capacityStr := c.QueryParam("capacity")
	pageSizeStr := c.QueryParam("pageSize")
	pageStr := c.QueryParam("page")

	// konversi query parameters ke tipe data yang sesuai
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
	var pageSize, page int
	if pageSizeStr != "" {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize <= 0 {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Message: "Invalid pageSize parameter",
			})
		}
	} else {
		pageSize = 10 // default page size
	}
	if pageStr != "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Message: "Invalid page parameter",
			})
		}
	} else {
		page = 1 // default page number
	}

	// hitung offset untuk pagination
	offset := (page - 1) * pageSize

	rooms, totalData, err := h.uc.RoomsList(roomName, roomType, capacity, pageSize, offset)
	if err != nil {
		switch err {
		case repository.ErrDatabase:
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: err.Error(),
			})
		case repository.ErrInternalServer:
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: err.Error(),
			})
		}
	}
	totalPage := (totalData + pageSize - 1) / pageSize // menghitung total halaman

	if len(rooms) == 0 {
		return c.JSON(http.StatusOK, utils.ListResponse{
			Message:   "No rooms found",
			Data:      []entity.Rooms{},
			Page:      page,
			PageSize:  pageSize,
			TotalPage: totalPage,
			TotalData: totalData,
		})
	}

	return c.JSON(http.StatusOK, utils.ListResponse{
		Message:   "Rooms list retrieved successfully",
		Data:      rooms,
		Page:      page,
		PageSize:  pageSize,
		TotalPage: totalPage,
		TotalData: totalData,
	})
}

// @Summary Create a new room
// @Description Create a new room (admin only)
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param room body request.CreateRoomRequest true "Room details"
// @Success 201 {object} utils.SuccessResponse{data=request.CreateRoomRequest} "Room created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request body"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Failure 409 {object} utils.ErrorResponse "Room already exists"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /rooms [post]
func (h *RoomsHandler) CreateRoom(c echo.Context) error {
	// ambil claim dari context
	claims := c.Get("client").(jwt.MapClaims)
	role, ok := claims["role"].(string)
	if !ok || role != "admin" {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid token claims",
		})
	}
	status, ok := claims["status"].(string)
	if !ok || status != "active" {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid token claims",
		})
	}

	var room request.CreateRoomRequest
	if err := c.Bind(&room); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request body",
		})
	}
	// validasi input
	if room.Name == "" || room.Type == "" || room.PricePerHour <= 0 || room.Capacity <= 0 || room.ImgUrl == "" {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Name, type, price per hour, and capacity are required and must be valid",
		})
	}

	imgUrl := room.ImgUrl
	if imgUrl != "" {
		domain := config.New().Domain

		// validasi url
		parsedURL, err := url.Parse(room.ImgUrl)
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
		data, err := UploadFile.UploadFile(room.ImgUrl)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Failed to upload file: " + err.Error(),
			})
		}

		// print log data
		fmt.Println("fileRequest: ", data)
		imgUrl = data.ImageURL
		room.ImgUrl = imgUrl
		fmt.Println("imgUrl: ", imgUrl)
	}

	err := h.uc.CreateRoom(&room, role, status)
	if err != nil {
		switch err {
		case repository.ErrForbidden:
			return c.JSON(http.StatusForbidden, utils.ErrorResponse{
				Message: err.Error(),
			})
		case repository.ErrRoomAlreadyExists:
			return c.JSON(http.StatusConflict, utils.ErrorResponse{
				Message: err.Error(),
			})
		case repository.ErrFailedToCreateRoom:
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: err.Error(),
			})
		}
	}

	return c.JSON(http.StatusCreated, utils.SuccessResponse{
		Message: "Room created successfully",
		Data:    room,
	})
}

// @Summary Update a room
// @Description Update a room by ID (admin only)
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Room ID"
// @Param room body request.CreateRoomRequest true "Updated room details"
// @Success 200 {object} utils.SuccessResponse{data=object} "Room updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request body or parameters"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Failure 404 {object} utils.ErrorResponse "Room not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /rooms/{id} [put]
func (h *RoomsHandler) UpdateRoom(c echo.Context) error {
	// ambil claim dari context
	claims := c.Get("client").(jwt.MapClaims)
	role, ok := claims["role"].(string)
	if !ok || role != "admin" {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid token claims",
		})
	}
	status, ok := claims["status"].(string)
	if !ok || status != "active" {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid token claims",
		})
	}

	// ambil id dari param
	id := c.Param("id")
	roomID, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid room ID",
		})
	}

	// ambil data dari body
	var room request.CreateRoomRequest
	if err := c.Bind(&room); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request body",
		})
	}
	// validasi input
	if room.Name == "" || room.Type == "" || room.PricePerHour <= 0 || room.Capacity <= 0 || room.ImgUrl == "" {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Name, type, price per hour, and capacity are required and must be valid",
		})
	}

	// ambil data imgUrl dari database berdasarkan id
	imgUrl, err := h.uc.GetImgRoomUrlByID(roomID)
	if err != nil {
		switch err {
		case repository.ErrRoomNotFound:
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Message: err.Error(),
			})
		case repository.ErrDatabase:
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: err.Error(),
			})
		}
	}

	// jika imgUrl dari body berbeda dengan imgUrl dari database, maka pindahkan file dari temp ke uploads
	if room.ImgUrl != imgUrl {
		domain := config.New().Domain

		// validasi url
		parsedURL, err := url.Parse(room.ImgUrl)
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
		data, err := UploadFile.UploadFile(room.ImgUrl)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Failed to upload file: " + err.Error(),
			})
		}

		// print log data
		fmt.Println("fileRequest: ", data)
		imgUrl = data.ImageURL
		room.ImgUrl = imgUrl
		fmt.Println("imgUrl: ", imgUrl)
	}

	err = h.uc.UpdateRoom(roomID, &room, role, status)
	if err != nil {
		switch err {
		case repository.ErrForbidden:
			return c.JSON(http.StatusForbidden, utils.ErrorResponse{
				Message: err.Error(),
			})
		case repository.ErrRoomNotFound:
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Message: err.Error(),
			})
		case repository.ErrFailedToUpdateRoom:
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: err.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Room updated successfully",
		Data:    nil,
	})
}

// @Summary Delete a room
// @Description Delete a room by ID (admin only)
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Room ID"
// @Success 200 {object} utils.SuccessResponse{data=object} "Room deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid room ID"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Failure 404 {object} utils.ErrorResponse "Room not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /rooms/{id} [delete]
func (h *RoomsHandler) DeleteRoom(c echo.Context) error {
	// ambil id dari param
	id := c.Param("id")
	roomID, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid room ID",
		})
	}

	// ambil claim dari context
	claims := c.Get("client").(jwt.MapClaims)
	role, ok := claims["role"].(string)
	if !ok || role != "admin" {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid token claims",
		})
	}
	status, ok := claims["status"].(string)
	if !ok || status != "active" {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid token claims",
		})
	}

	err = h.uc.DeleteRoom(roomID)
	if err != nil {
		switch err {
		case repository.ErrFailedToDeleteRoom:
			return c.JSON(http.StatusForbidden, utils.ErrorResponse{
				Message: "Room cannot be deleted as it is associated with existing transactions",
			})
		case repository.ErrRoomNotFound:
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Message: err.Error(),
			})
		case repository.ErrInternalServer:
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: err.Error(),
			})
		}
	}

	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Room deleted successfully",
		Data:    nil,
	})
}
