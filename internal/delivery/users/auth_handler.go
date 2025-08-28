package delivery

import (
	"e_meeting/internal/entity"
	"e_meeting/internal/models/request"
	"e_meeting/internal/models/response"
	usecase "e_meeting/internal/usecase/users"
	"e_meeting/pkg/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UsersHandler struct {
	uc *usecase.AuthUsecase
}

func NewUsersHandler(e *echo.Echo, uc *usecase.AuthUsecase) {
	handler := &UsersHandler{uc}
	e.POST("/register", handler.Register)
	e.POST("/login", handler.Login)
	e.POST("/refresh_token", handler.ResetAccessToken)
}

// @Summary Register a new user
// @Description Register a new user
// @Tags authentication
// @Accept json
// @Produce json
// @Param user body request.RegisterRequest true "User registration data"
// @Success 200 {object} utils.SuccessResponse{data=nil}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /register [post]
func (h *UsersHandler) Register(c echo.Context) error {
	// bind request data
	var reqReg request.RegisterRequest
	if err := c.Bind(&reqReg); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request data : " + err.Error(),
		})
	}

	// data request untuk register
	user := &entity.Users{
		Username: reqReg.Username,
		Email:    reqReg.Email,
		Password: reqReg.Password,
	}

	// panggil usecase register
	if err := h.uc.Register(user, reqReg.ConfirmPass); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "User registered successfully",
	})

}

// @Summary Login a user
// @Description Login a user
// @Tags authentication
// @Accept json
// @Produce json
// @Param user body request.LoginRequest true "User login data"
// @Success 200 {object} response.LoginResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /login [post]
func (h *UsersHandler) Login(c echo.Context) error {
	// bind request data
	var reqLogin request.LoginRequest
	if err := c.Bind(&reqLogin); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request data : " + err.Error(),
		})
	}

	// panggil usecase login
	accessToken, refreshToken, err := h.uc.Login(reqLogin.Username, reqLogin.Password)
	if err != nil {
		return c.JSON(http.StatusNotFound, utils.ErrorResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, response.LoginResponse{
		Message: "User logged in successfully",
		Data: response.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	})
}

// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags authentication
// @Accept json
// @Produce json
// @Param request body request.RefreshTokenRequest true "Refresh token data"
// @Success 200 {object} utils.SuccessResponse{data=string}
// @Failure 400 {object} utils.ErrorResponse
// @Router /refresh_token [post]
func (h *UsersHandler) ResetAccessToken(c echo.Context) error {
	// bind request data
	var req request.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request data : " + err.Error(),
		})
	}
	newAccessToken, err := h.uc.RefreshAccessToken(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Access token refreshed successfully",
		Data:    newAccessToken,
	})
}
