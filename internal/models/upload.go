package models

type UploadRequest struct {
	ImageURL string `json:"imageUrl"`
}

type UploadResult struct {
	Data UploadRequest
	Err  error
}
