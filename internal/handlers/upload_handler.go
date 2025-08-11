package handlers

import (
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func InitUploadHandler(e *echo.Group) {
	e.POST("/upload", func(c echo.Context) error {
		return Upload(c)
	})
	e.POST("/upload/file", func(c echo.Context) error {
		return UploadFile(c, make(chan models.UploadRequest), "")
	})
}

// @Summary Upload an image
// @Description Upload an image to the server
// @Tags upload image
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Image file"
// @Success 200 {object} utils.SuccessResponse{data=nil}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /upload [post]
func Upload(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid request data : " + err.Error(),
		})
	}

	// fmt.Println("Uploaded file:", file.Filename)
	// fmt.Println("File size in megabytes:", file.Size)
	// validasi file extension harus jpg, jpeg, png
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "File extension must be .jpg, .jpeg, or .png",
		})
	}
	// validasi file size agar tidak melebihi 1MB
	if file.Size > 1024*1024 {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "File size must be less than 1MB",
		})
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid open file : " + err.Error(),
		})
	}
	defer src.Close()

	// Destination
	dst, err := os.Create("temp/" + file.Filename)
	if err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid create file : " + err.Error(),
		})
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return c.JSON(http.StatusBadRequest, utils.ErrorResponse{
			Message: "Invalid copy file : " + err.Error(),
		})
	}
	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "success upload file to temp",
		Data:    file.Filename,
	})
}

func UploadFile(c echo.Context, ch chan models.UploadRequest, imgUrl string) error {
	request := models.UploadRequest{
		ImageURL: imgUrl,
	}

	srcPath := "temp/" + request.ImageURL                   // path file temp
	ext := filepath.Ext(srcPath)                            // Mendapatkan ekstensi file
	name := strings.TrimSuffix(filepath.Base(srcPath), ext) // Mendapatkan nama file

	// Tambahkan timestamp
	timestamp := time.Now().Format("20060102150405")            // YYYYMMDDHHMMSS
	newFilename := fmt.Sprintf("%s_%s%s", name, timestamp, ext) // Nama file baru

	dstPath := filepath.Join("uploads", newFilename) // path file uploads

	// Pastikan folder uploads ada
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to create uploads folder: " + err.Error(),
		})
	}

	// Pindahkan file dari temp ke uploads
	if err := os.Rename(srcPath, dstPath); err != nil {
		return c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
			Message: "Failed to move file: " + err.Error(),
		})
	}

	// Kirim data path uploads ke channel
	ch <- models.UploadRequest{
		ImageURL: dstPath,
	}

	fmt.Println("File uploaded:", dstPath)

	return nil
}
