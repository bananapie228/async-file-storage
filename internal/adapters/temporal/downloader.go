package temporaladapter

import (
	"context"
	"fmt"
	"time"

	"async-file-storage/internal/temporal"

	"go.temporal.io/sdk/client"
)

type Downloader struct {
	client    client.Client
	taskQueue string
}

func NewDownloader(c client.Client, taskQueue string) *Downloader {
	return &Downloader{client: c, taskQueue: taskQueue}
}

func (d *Downloader) StartDownload(ctx context.Context, requestID int, urls []string, timeout time.Duration) error {
	options := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("download_request_%d", requestID),
		TaskQueue: d.taskQueue,
	}

	_, err := d.client.ExecuteWorkflow(ctx, options, temporal.DownloadWorkflow, requestID, urls, timeout)
	if err != nil {
		return fmt.Errorf("execute workflow: %w", err)
	}
	return nil
}
