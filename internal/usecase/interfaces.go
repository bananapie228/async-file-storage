package usecase

import (
	"context"
	"time"

	"async-file-storage/internal/domain"
)

type Repository interface {
	CreateRequest(ctx context.Context, urls []string) (int, error)
	GetRequestStatus(ctx context.Context, id int) (*domain.DownloadRequest, []domain.FileEntry, error)
	GetFile(ctx context.Context, requestID int, fileID int) (*domain.FileEntry, error)
}

type Downloader interface {
	StartDownload(ctx context.Context, requestID int, urls []string, timeout time.Duration) error
}
