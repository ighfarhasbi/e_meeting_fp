package delivery

import (
	usecase "e_meeting/internal/usecase/snacks"
	"e_meeting/pkg/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type SnacksHandler struct {
	uc *usecase.SnacksUsecase
}

func NewSnackHandler(e *echo.Group, uc *usecase.SnacksUsecase) {
	handler := &SnacksHandler{uc}
	e.GET("/snacks", handler.SnacksList)
}

// @Summary Get snacks list
// @Description Retrieve a list of all available snacks
// @Tags snacks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.SuccessResponse{data=[]entity.Snacks}
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /snacks [get]
func (h *SnacksHandler) SnacksList(c echo.Context) error {
	snackList, err := h.uc.SnacksList()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Snacks list retrieved successfully",
		Data:    snackList,
	})
}
