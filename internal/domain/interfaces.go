package domain

import "context"

type Storage interface {
	CreateRequest(ctx context.Context, urls []string) (int, error)
	UpdateRequestStatus(ctx context.Context, id int, status Status) error
	UpdateFileStatus(ctx context.Context, requestID int, url string, data []byte, downloadErr error) error
	GetRequestStatus(ctx context.Context, id int) (*DownloadRequest, []FileEntry, error)
}
