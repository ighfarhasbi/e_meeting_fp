package handlers

import (
	"database/sql"
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"net/http"

	"github.com/labstack/echo"
)

// InitUserHandler initializes the user handler
func InitUserHandler(e *echo.Echo, dbConn *sql.DB) {
	e.GET("/users/:user_id", func(c echo.Context) error {
		return GetUserById(c, dbConn)
	})
}

func GetUserById(c echo.Context, db *sql.DB) error {
	// ambil user_id dari parameter
	userId := c.Param("user_id")

	// query untuk mengambil data user berdasarkan user_id
	row := db.QueryRow("SELECT users_id, username, email, role, status, language, img_path, created_at, updated_at FROM users WHERE users_id = $1", userId)

	// buat struct untuk menyimpan data user
	var user models.Users
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.Status, &user.Language, &user.ImgPath, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Message: "User not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to retrieve user",
		})
	}

	// kembalikan data user sebagai response
	return c.JSON(http.StatusOK, utils.Response{
		Message: "User retrieved successfully",
		Data:    user,
	})
}
