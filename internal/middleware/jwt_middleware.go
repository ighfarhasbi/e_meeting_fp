package middleware

import (
	"e_meeting/pkg/utils"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func JwtMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Ambil token dari header Authorization
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Message: "Authorization header is missing",
			})
		}

		// Validasi token JWT
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Message: "Bearer token is missing",
			})
		}

		// Verifikasi token
		if _, err := utils.VerifyToken(tokenString); err != nil {
			return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Message: "Invalid or expired token : " + err.Error(),
			})
		}

		return next(c)
	}
}
