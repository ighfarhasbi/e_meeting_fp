package handlers

import (
	"e_meeting/config"
	"e_meeting/internal/models"
	"e_meeting/pkg/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func InitUploadHandler(e *echo.Group) {
	e.POST("/upload", func(c echo.Context) error {
		return Upload(c)
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

	// tambahkan timestamp di filename
	name := strings.TrimSuffix(file.Filename, ext)
	timestamp := time.Now().Format("20060102150405") // YYYYMMDDHHMMSS
	newFilename := fmt.Sprintf("%s_%s%s", name, timestamp, ext)

	// Destination
	dst, err := os.Create("temp/" + newFilename)
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

	// tambahkan nama domain di filename
	domain := config.New().Domain
	fileName := domain + "/temp/" + newFilename

	return c.JSON(http.StatusOK, utils.SuccessResponse{
		Message: "success upload file to temp",
		Data:    fileName,
	})
}

func UploadFile(c echo.Context, imgUrl string) (models.UploadRequest, error) {
	request := models.UploadRequest{
		ImageURL: imgUrl,
	}

	// ambil url dari .env untuk path file
	// cfg := config.New()
	// domain := cfg.Domain

	// ambil fileName dari request
	fileName := path.Base(request.ImageURL)

	srcPath := "temp/" + fileName // path file temp
	// validasi fileName sama dengan yang ada di folder temp
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return models.UploadRequest{ImageURL: ""}, fmt.Errorf("file not found in temp folder")
	} else if err != nil {
		return models.UploadRequest{ImageURL: ""}, fmt.Errorf("error checking file: %v", err)
	}

	dstPath := filepath.Join("uploads", fileName)

	// Pastikan folder uploads ada
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		return models.UploadRequest{ImageURL: ""}, fmt.Errorf("error creating uploads folder: %v", err)
	}

	// Pindahkan file dari temp ke uploads
	if err := os.Rename(srcPath, dstPath); err != nil {
		return models.UploadRequest{ImageURL: ""}, fmt.Errorf("error moving file: %v", err)
	}

	// setelah file dipindahkan, tambahkan domain disini
	// publicURL := fmt.Sprintf("%s/%s", domain, dstPath)

	fmt.Println("File uploaded:", dstPath)

	return request, nil
}
