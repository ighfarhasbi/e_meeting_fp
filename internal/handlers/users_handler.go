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
	e.POST("/register", func(c echo.Context) error {
		return RegisterUser(c, dbConn)
	})
	e.POST("/login", func(c echo.Context) error {
		return LoginUser(c, dbConn)
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
			Message: "Failed to retrieve user : " + err.Error(),
		})
	}

	// kembalikan data user sebagai response
	return c.JSON(http.StatusOK, utils.Response{
		Message: "User retrieved successfully",
		Data:    user,
	})
}

func RegisterUser(c echo.Context, db *sql.DB) error {
	var user models.RegisterUserRequest
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request data : " + err.Error(),
		})
	}

	// hashing password with bcrypt
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to hash password : " + err.Error(),
		})
	}
	// insert user into database
	_, err = db.Exec("INSERT INTO users (username, email, password) VALUES ($1, $2, $3)", user.Username, user.Email, hashedPassword)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to register user : " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, utils.RegisterResposnse{
		Message: "User registered successfully",
	})
}

func LoginUser(c echo.Context, db *sql.DB) error {
	// bind request data
	var loginRequest models.LoginUserRequest
	if err := c.Bind(&loginRequest); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request data : " + err.Error(),
		})
	}

	// query untuk mengambil data user berdasarkan username
	row := db.QueryRow("SELECT users_id, username, email, password, role, status, language, img_path, created_at, updated_at FROM users WHERE username = $1", loginRequest.Username)

	// buat struct untuk menyimpan data user
	var hashedPassword string
	var user models.Users
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &hashedPassword, &user.Role, &user.Status, &user.Language, &user.ImgPath, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Message: "Invalid username or password",
			})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to retrieve user : " + err.Error(),
		})
	}

	// cek password
	if utils.ValidatePassword(hashedPassword, loginRequest.Password) != nil {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid username or password",
		})
	}

	// generate JWT tokens (access and refresh)
	accessToken, refreshToken, err := utils.GenerateJWTToken(user.Username, user.Role, user.Status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to generate tokens : " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, utils.LoginResponse{
		Message:      "Login successful",
		Data:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
