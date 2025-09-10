package delivery

import (
	"e_meeting/internal/entity"
	repository "e_meeting/internal/repository/rooms"
	usecase "e_meeting/internal/usecase/rooms"
	"e_meeting/pkg/utils"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type RoomsHandler struct {
	uc *usecase.RoomsUsecase
}

func NewRoomHandler(e *echo.Group, uc *usecase.RoomsUsecase) {
	handler := &RoomsHandler{uc}
	e.GET("/rooms", handler.RoomsList)
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
