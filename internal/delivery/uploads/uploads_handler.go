package delivery

import (
	usecase "e_meeting/internal/usecase/uploads"
	"e_meeting/pkg/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UploadsHandler struct {
	uc usecase.UploadUsecase
}

func NewUploadsHandler(e *echo.Group, uc usecase.UploadUsecase) {
	handler := &UploadsHandler{uc}
	e.POST("/upload", handler.Upload)
}

// @Summary Upload an image
// @Description Upload an image to the server
// @Tags upload image
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Image file"
// @Success 200 {object} utils.SuccessResponse{data=nil}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /upload [post]
func (h *UploadsHandler) Upload(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request data : " + err.Error(),
		})
	}
	src, _ := file.Open()
	defer src.Close()

	url, err := h.uc.SaveTemp(file.Filename, file.Size, src)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "success upload file to temp",
		Data:    url,
	})
}
