package usecase

import (
	repository "e_meeting/internal/repository/uploads"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// interface usecase
// type UploadUsecase interface {
// 	SaveTemp(fileName string, size int64, src io.Reader) (string, error)
// 	MoveToUploads(fileName string) (string, error)
// }

// implementasi usecase
type uploadUsecase struct {
	domain string
}

// usecase/upload_service.go
// type UploadUsecase struct {
// 	Domain string
// }

func NewUploadUsecase(repo repository.UploadsRepository, domain string) UploadUsecase {
	return &uploadUsecase{
		domain: domain,
	}
}

func (uc *uploadUsecase) SaveTemp(fileName string, size int64, src io.Reader) (string, error) {
	// validasi ekstensi file dan ukuran file
	ext := filepath.Ext(fileName)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", fmt.Errorf("file extension must be .jpg, .jpeg, or .png")
	}
	if size > 1024*1024 {
		return "", fmt.Errorf("file size must be less than 1MB")
	}

	// buat folder temp jika belum ada
	err := os.MkdirAll("temp", os.ModePerm)
	if err != nil {
		return "", err
	}

	// rename file dengan menambahkan timestamp
	newName := fmt.Sprintf("%s_%s%s",
		strings.TrimSuffix(fileName, filepath.Ext(fileName)),
		time.Now().Format("20060102150405"),
		filepath.Ext(fileName),
	)

	dst, err := os.Create("temp/" + newName)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return uc.domain + "/temp/" + newName, nil
}

func (uc *uploadUsecase) MoveToUploads(fileName string) (string, error) {
	srcPath := "temp/" + fileName
	dstPath := "uploads/" + fileName

	// buat folder uploads jika belum ada
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		return "", err
	}
	if err := os.Rename(srcPath, dstPath); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/uploads/%s", uc.domain, fileName), nil
}
