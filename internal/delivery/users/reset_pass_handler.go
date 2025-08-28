package delivery

import (
	"e_meeting/config"
	"e_meeting/internal/models/request"
	usecase "e_meeting/internal/usecase/users"
	"e_meeting/pkg/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ResetPassHandler struct {
	uc *usecase.ResetPassUsecase
}

func NewResetPassHandler(e *echo.Echo, uc *usecase.ResetPassUsecase) {
	handler := &ResetPassHandler{uc}
	e.POST("/password/reset_request", handler.CheckEmailExists)
	e.PUT("/password/reset/:token", handler.ResetPassword)
}

// @Summary Check email exists
// @Description Check if an email already exists in the database
// @Tags reset password
// @Accept json
// @Produce json
// @Param email body request.CheckEmailRequest true "Email data"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /password/reset_request [post]
func (h *ResetPassHandler) CheckEmailExists(c echo.Context) error {
	// bind request data
	var req request.CheckEmailRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request data : " + err.Error(),
		})
	}
	if err := h.uc.CheckEmailExists(req.Email); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Email sent successfully",
	})
}

// @Summary Reset password
// @Description Reset user password using reset token
// @Tags reset password
// @Accept json
// @Produce json
// @Param id path string true "Token reset password"
// @Param request body request.UpdatePassRequest true "New password data"
// @Success 200 {object} utils.SuccessResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /password/reset/{id} [put]
func (h *ResetPassHandler) ResetPassword(c echo.Context) error {
	// ambil secret key dari config
	cgf := config.New()
	secKey := cgf.JWTResetPassword
	// ambil token dari path parameter
	token := c.Param("token")
	// bind request data
	var req request.UpdatePassRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request data : " + err.Error(),
		})
	}
	if err := h.uc.ResetPassword(token, secKey, req); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Password reset successfully",
	})
}
