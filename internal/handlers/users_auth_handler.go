package handlers

import (
	"database/sql"
	"e_meeting/config"
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// InitUserHandler initializes the user handler
func InitUserAuthHandler(e *echo.Echo, dbConn *sql.DB) {
	e.POST("/register", func(c echo.Context) error {
		return RegisterUser(c, dbConn)
	})
	e.POST("/login", func(c echo.Context) error {
		return LoginUser(c, dbConn)
	})
	e.POST("/refresh_token", func(c echo.Context) error {
		return RefreshAccessToken(c)
	})
	e.POST("/password/reset_request", func(c echo.Context) error {
		return CheckEmailExists(c, dbConn)
	})
	e.PUT("/password/reset/:id", func(c echo.Context) error {
		return ResetPassword(c, dbConn)
	})
}

// Summary RegisterUser handles user registration
// Description Register a new user
// Tags authentication
// Accept json
// Produce json
// Param user body models.RegisterUserRequest true "User registration data"
// Success 200 {object} utils.SuccessResponse{data=nil}
// Failure 400 {object} utils.ErrorResponse
// Failure 500 {object} utils.ErrorResponse
// Router /register [post]
func RegisterUser(c echo.Context, db *sql.DB) error {
	var user models.RegisterUserRequest
	if err := c.Bind(&user); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request data : " + err.Error(),
		})
	}

	// validasi password characters
	if err := utils.ValidatePasswordCharacters(user.Password); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Password validation failed : " + err.Error(),
		})
	}
	// validasi password
	if user.Password != user.ConfirmPassword {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Password and confirm password do not match",
		})
	}

	// validasi email
	if err := utils.ValidateEmail(user.Email); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Email validation failed : " + err.Error(),
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

	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "User registered successfully",
		Data:    nil,
	})
}

// Summary LoginUser handles user login
// Description Authenticate user and return JWT tokens
// Tags authentication
// Accept json
// Produce json
// Param user body models.LoginUserRequest true "User login data"
// Success 200 {object} utils.SuccessResponse{data=utils.TokenResponse}
// Failure 400 {object} utils.ErrorResponse
// Failure 500 {object} utils.ErrorResponse
// Router /login [post]
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
			return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Message: "Invalid username or password",
			})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to retrieve user : " + err.Error(),
		})
	}

	// cek password
	if utils.ValidatePassword(hashedPassword, loginRequest.Password) != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid username or password",
		})
	}

	// generate JWT tokens (access and refresh)
	accessToken, refreshToken, err := utils.GenerateJWTToken(user.ID, user.Username, user.Role, user.Status)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to generate tokens : " + err.Error(),
		})
	}
	// buat struct untuk menyimpan data token
	tokenResponse := utils.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Login successful",
		Data:    tokenResponse,
	})
}

// @Summary RefreshAccessToken handles token refresh
// @Description Refresh the access token using the refresh token
// @Tags authentication
// @Accept json
// @Produce json
// @Param refresh_token body models.RefreshTokenRequest true "Refresh token data"
// @Success 200 {object} utils.SuccessResponse{data=models.AccessToken}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /refresh_token [post]
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

// @Summary CheckEmailExists checks if an email already exists in the database
// @Description Check if an email already exists in the database
// @Tags reset password
// @Accept json
// @Produce json
// @Param email body models.CheckEmailRequest true "Email data"
// @Success 200 {object} utils.RegisterResposnse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.RegisterResposnse
// @Failure 500 {object} utils.ErrorResponse
// @Router /password/reset_request [post]
func CheckEmailExists(c echo.Context, db *sql.DB) error {
	var request models.CheckEmailRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request data : " + err.Error(),
		})
	}

	var user models.Users
	// query untuk memeriksa apakah email sudah ada di database
	row := db.QueryRow("SELECT users_id, username, email, role, status, language, img_path, created_at, updated_at FROM users WHERE email = $1", request.Email)
	if err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Role, &user.Status, &user.Language, &user.ImgPath, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, utils.RegisterResposnse{
				Message: "Email not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to check email existence : " + err.Error(),
		})
	}

	// jika email ditemukan, kirim email reset password menggunakan smtp
	token, err := utils.GenerateResetToken(user.Email, user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to generate reset token : " + err.Error(),
		})
	}
	// kirim email reset password dengan goroutine
	go func(email, token string) {
		if err := utils.SendEmail(email, token); err != nil {
			log.Printf("[SendEmail Error] to %s: %v", email, err)
		}
	}(user.Email, token)

	// tanpa goroutine, langsung kirim email
	// if err := utils.SendEmail(user.Email, token); err != nil {
	// 	return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
	// 		Message: "Failed to send reset password email : " + err.Error(),
	// 	})
	// }

	// masukkan token ke dalam database untuk reset password
	_, err = db.Exec("INSERT INTO password_resets (token) VALUES ($1)", token)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to save reset token : " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "Email existence checked successfully, reset password email sent",
		Data:    user,
	})
}

// @Summary ResetPassword handles password reset
// @Description Reset user password using reset token
// @Tags reset password
// @Accept json
// @Produce json
// @Param id path string true "Reset token"
// @Param request body models.ResetPasswordRequest true "New password data"
// @Success 200 {object} utils.RegisterResposnse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /password/reset/{id} [put]
func ResetPassword(c echo.Context, db *sql.DB) error {
	cfg := config.New() // Ambil JWT secret dari konfigurasi
	// ambil path parameter token
	id := c.Param("id")
	// ambil body request untuk password baru
	var request models.ResetPasswordRequest
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request data : " + err.Error(),
		})
	}

	// validasi password characters
	if err := utils.ValidatePasswordCharacters(request.NewPassword); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Password validation failed : " + err.Error(),
		})
	}
	// validasi password baru
	if request.NewPassword != request.ConfirmPassword {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "New password and confirm password do not match",
		})
	}
	// ambil token dari database berdasarkan token yang diberikan dari path parameter
	var resetToken models.PasswordReset
	row := db.QueryRow("SELECT id, token FROM password_resets WHERE token = $1", id)
	if err := row.Scan(&resetToken.ID, &resetToken.Token); err != nil {
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusNotFound, utils.ErrorResponse{
				Message: "Reset token not found: " + resetToken.Token,
			})
		}
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to retrieve reset token : " + err.Error(),
		})
	}
	// validasi token
	token, err := jwt.Parse(resetToken.Token, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(cfg.JWTResetPassword), nil
	})
	if err != nil || !token.Valid {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid reset token",
		})
	}
	// ambil user_id dari klaim token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid token claims",
		})
	}
	userID := claims["user_id"].(float64)

	// hash password baru
	hashedPassword, err := utils.HashPassword(request.NewPassword)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to hash new password : " + err.Error(),
		})
	}

	// implementasi transaksi untuk update password
	tx, err := db.Begin()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to begin transaction : " + err.Error(),
		})
	}
	defer tx.Rollback() // rollback jika terjadi error
	// update password di database
	_, err = tx.Exec("UPDATE users SET password = $1 WHERE users_id = $2", hashedPassword, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to update password : " + err.Error(),
		})
	}
	// hapus token reset password dari database
	_, err = tx.Exec("DELETE FROM password_resets WHERE id = $1", resetToken.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to delete reset token : " + err.Error(),
		})
	}
	// commit transaksi
	if err := tx.Commit(); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to commit transaction : " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, utils.RegisterResposnse{
		Message: "Password reset successfully",
	})
}
