package domain

import "time"

type Status string

const (
	StatusProcess Status = "PROCESS"
	StatusDone    Status = "DONE"
	StatusError   Status = "ERROR"
)

type DownloadRequest struct {
	ID        int
	Status    Status
	CreatedAt time.Time
}

type FileEntry struct {
	ID        int
	RequestID int
	URL       string
	Data      []byte
	Error     string
}
