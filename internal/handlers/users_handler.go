package handlers

import (
	"database/sql"
	"e_meeting/internal/middleware"
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// InitUserHandler initializes the user handler
func InitUserHandler(e *echo.Echo, dbConn *sql.DB) {
	// sisipkan middleware jwt pada salah satu endpoint
	group := e.Group("/users")
	group.Use(middleware.JwtMiddleware)
	// inisialisasi endpoint untuk user
	group.GET("/:user_id", func(c echo.Context) error {
		return GetUserById(c, dbConn)
	})
	// e.GET("/users/:user_id", func(c echo.Context) error {
	// 	return GetUserById(c, dbConn)
	// })
	e.POST("/register", func(c echo.Context) error {
		return RegisterUser(c, dbConn)
	})
	e.POST("/login", func(c echo.Context) error {
		return LoginUser(c, dbConn)
	})
	e.POST("/refresh-token", func(c echo.Context) error {
		return RefreshAccessToken(c)
	})
}

// @Summary GetUserById retrieves user data by user_id
// @Description Get user data by user_id
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID"
// @Success 200 {object} utils.SuccessResponse{data=models.Users}
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /users/{user_id} [get]
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
	// userId := c.Param("user_id")

	// query untuk mengambil data user berdasarkan user_id
	// row := db.QueryRow("SELECT users_id, username, email, role, status, language, img_path, created_at, updated_at FROM users WHERE users_id = $1", userId)
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

	// kembalikan data user sebagai response
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "User retrieved successfully",
		Data:    user,
	})
}

// @Summary RegisterUser handles user registration
// @Description Register a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.RegisterUserRequest true "User registration data"
// @Success 200 {object} utils.RegisterResposnse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /register [post]
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

// @Summary LoginUser handles user login
// @Description Authenticate user and return JWT tokens
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.LoginUserRequest true "User login data"
// @Success 200 {object} utils.LoginResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /login [post]
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

// @Summary RefreshAccessToken handles token refresh
// @Description Refresh the access token using the refresh token
// @Tags users
// @Accept json
// @Produce json
// @Param refresh_token body models.RefreshTokenRequest true "Refresh token data"
// @Success 200 {object} utils.SuccessResponse{data=models.AccessToken}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /refresh-token [post]
func RefreshAccessToken(c echo.Context) error {
	// ambil refresh token dari body
	var request models.RefreshTokenRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request data : " + err.Error(),
		})
	}
	// refresh access token
	newAccessToken, err := utils.RefreshAccessToken(request.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Failed to refresh access token : " + err.Error(),
		})
	}
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Access token refreshed successfully",
		Data:    models.AccessToken{AccessToken: newAccessToken},
	})
}
