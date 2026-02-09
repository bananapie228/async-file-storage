package usecase

import (
	"time"

	"async-file-storage/internal/domain"
)

type CreateRequestInput struct {
	URLs    []string
	Timeout time.Duration
}

type CreateRequestOutput struct {
	ID     int
	Status domain.Status
}

type GetRequestOutput struct {
	ID     int
	Status domain.Status
	Files  []FileStatus
}

type FileStatus struct {
	URL       string
	FileID    int
	ErrorCode string
}

type GetFileOutput struct {
	Data []byte
}
