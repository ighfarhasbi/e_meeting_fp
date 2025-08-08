package middlewareAuth

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
		token, err := utils.VerifyToken(tokenString)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Message: "Invalid or expired token: " + err.Error(),
			})
		}

		// masukkan token ke dalam context untuk digunakan di handler
		c.Set("client", token.Claims)

		return next(c)
	}
}
