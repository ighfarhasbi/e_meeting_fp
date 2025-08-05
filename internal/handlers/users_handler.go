package handlers

import (
	"database/sql"
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// InitUserHandler initializes the user handler
func InitUserHandler(e *echo.Group, dbConn *sql.DB) {
	// // sisipkan middleware jwt pada salah satu endpoint
	// group := e.Group("/users")
	// group.Use(middleware.JwtMiddleware)
	// inisialisasi endpoint untuk user
	e.GET("/users/:id", func(c echo.Context) error {
		return GetUserById(c, dbConn)
	})
	e.PUT("/users/:id", func(c echo.Context) error {
		return EditProfile(c, dbConn)
	})
}

// @Summary GetUserById retrieves user data by user_id
// @Description Get user data by user_id
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} utils.SuccessResponse{data=models.Users}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /users/{id} [get]
func GetUserById(c echo.Context, db *sql.DB) error {
	// ambil header Authorization
	authHeader := c.Request().Header.Get("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	// parse token untuk mendapatkan username
	token, _ := utils.VerifyToken(tokenStr)
	// ambil username dari klaim token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["username"] == nil {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid token claims",
		})
	}
	username := claims["username"].(string)

	// ambil user_id dari parameter
	userId := c.Param("id")
	// konfersi user_id ke int menggunakn strconv.Atoi
	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid id format : " + err.Error(),
		})
	}
	// query untuk mengambil data user berdasarkan username
	row := db.QueryRow("SELECT users_id, username, email, role, status, language, img_path, created_at, updated_at FROM users WHERE username = $1", username)

	// buat struct untuk menyimpan data user
	var user models.Users
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.Status, &user.Language, &user.ImgPath, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Message: "User not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to retrieve user : " + err.Error(),
		})
	}
	// bandingkan user_id dari parameter dengan user_id dari token
	if user.ID != userIdInt {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Unauthorized access to user data",
		})
	}

	// kembalikan data user sebagai response
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// @Summary EditProfile handles user profile editing
// @Description Edit user profile data
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body models.UpdateUserRequest true "User profile data"
// @Success 200 {object} utils.SuccessResponse{data=models.Users}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /users/{id} [put]
func EditProfile(c echo.Context, db *sql.DB) error {
	// ambil parameter user_id dari path
	id := c.Param("id")
	// konversi id ke int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid user_id : " + err.Error(),
		})
	}
	// ambil header Authorization
	authHeader := c.Request().Header.Get("Authorization")
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	// parse token untuk mendapatkan user_id
	token, err := utils.VerifyToken(tokenStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid token : " + err.Error(),
		})
	}
	// ambil user_id dari klaim token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["id"] == nil {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid token claims",
		})
	}
	userID := int(claims["id"].(float64))
	// bandingkan user_id dari parameter dengan user_id dari token
	if userID != idInt {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Unauthorized access to user data",
		})
	}
	// ambil request body
	var request models.UpdateUserRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request body : " + err.Error(),
		})
	}

	// list error
	var validationErrors []string
	// validasi email
	if err := utils.ValidateEmail(request.Email); err != nil {
		validationErrors = append(validationErrors, "Email validation failed: "+err.Error())
	}
	// validasi username apabila kosong
	if request.Username == "" {
		validationErrors = append(validationErrors, "Username cannot be empty")
	}
	// return error jika ada error validasi dalam bentuk array
	if len(validationErrors) > 0 {
		return c.JSON(http.StatusBadRequest, utils.MultupleErrorResponse{
			Errors: validationErrors,
		})
	}

	var user models.Users
	// update data user di database dengan transaction
	tx, err := db.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to start transaction : " + err.Error(),
		})
	}
	defer tx.Rollback() // rollback jika terjadi error
	_, err = tx.Exec("UPDATE users SET username = $1, email = $2, language = $3, img_path = $4 WHERE users_id = $5",
		request.Username, request.Email, request.Language, request.ImgPath, userID)
	// ambil data user yang sudah diupdate
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to update user : " + err.Error(),
		})
	}
	row := tx.QueryRow("SELECT users_id, username, email, role, status, language, img_path, created_at, updated_at FROM users WHERE users_id = $1", userID)
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.Status, &user.Language, &user.ImgPath, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Message: "User not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to retrieve updated user : " + err.Error(),
		})
	}
	// commit transaction
	if err := tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to commit transaction : " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "User profile updated successfully",
		Data:    user,
	})
}
