package handlers

import (
	"database/sql"
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"fmt"
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
	// ambil claim token dari context
	claims := c.Get("client").(jwt.MapClaims)
	// ambil user_id dari klaim token
	userIDfloat, ok := claims["id"].(float64)
	if !ok {
		return c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
			Message: "Invalid token claims",
		})
	}
	userID := int(userIDfloat) // konversi ke int

	// ambil parameter user_id dari path
	id := c.Param("id")
	// konversi id ke int
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid user_id : " + err.Error(),
		})
	}

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
	// cek apakah request.ImgPath sama dengan yang ada di database
	imgrow := db.QueryRow("SELECT img_path FROM users WHERE users_id = $1", userID)
	var dbImgPath string
	if err := imgrow.Scan(&dbImgPath); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to get user image path : " + err.Error(),
		})
	}
	imgUrl := dbImgPath
	if dbImgPath != request.ImgPath {
		// buat channel untuk menerima data string dan err dari func UploadFile
		ch := make(chan models.UploadRequest)
		defer close(ch)

		// jalankan goroutine untuk upload file
		go UploadFile(c, ch, request.ImgPath)

		// ambil data dari channel
		fileRequest := <-ch
		fmt.Println("fileRequest ch: ", fileRequest.ImageURL)
		imgUrl = fileRequest.ImageURL
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
	// validasi password jika ada
	if request.Password != "" {
		if err := utils.ValidatePasswordCharacters(request.Password); err != nil {
			validationErrors = append(validationErrors, "Password validation failed: "+err.Error())
		}
	}
	// return error jika ada error validasi dalam bentuk array
	if len(validationErrors) > 0 {
		return c.JSON(http.StatusBadRequest, utils.MultipleErrorResponse{
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
	// query untuk jika request.password tidak kosong
	if request.Password != "" {
		// hash password
		hashedPassword, err := utils.HashPassword(request.Password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Failed to hash password : " + err.Error(),
			})
		}
		// update query
		_, err = tx.Exec("UPDATE users SET username = $1, email = $2, language = $3, img_path = $4, password = $5 WHERE users_id = $6",
			request.Username, request.Email, request.Language, imgUrl, hashedPassword, userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Failed to update password : " + err.Error(),
			})
		}
	} else if request.Password == "" {
		// update query
		_, err = tx.Exec("UPDATE users SET username = $1, email = $2, language = $3, img_path = $4 WHERE users_id = $5",
			request.Username, request.Email, request.Language, imgUrl, userID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Message: "Failed to update user : " + err.Error(),
			})
		}
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
