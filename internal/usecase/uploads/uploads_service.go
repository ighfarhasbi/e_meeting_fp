package usecase

import "io"

type UploadsRepository interface {
	SaveTemp(fileName string, size int64, src io.Reader) (string, error)
	MoveToUploads(fileName string) (string, error)
}
