package temporal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"async-file-storage/internal/domain"

	"golang.org/x/sync/errgroup"
)

type Activities struct {
	Repo domain.Storage
}

// download multiple files and save to the DB
func (a *Activities) DownloadFilesActivity(ctx context.Context, RequestID int, urls []string) ([]string, error) {
	// Create an errgroup to limit concurrent downloads
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(3)

	results := make([]string, len(urls))

	for i, url := range urls {
		index := i
		link := url

		g.Go(func() error {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			fmt.Printf("[%d] downloading: %s/n ", index, link)

			data, downloadErr := downloadHelper(ctx, link)

			dbErr := a.Repo.UpdateFileStatus(ctx, RequestID, link, data, downloadErr)

			if dbErr != nil {
				fmt.Printf("downloading error: %v/n,", dbErr)
				return fmt.Errorf("database error: %w/n", dbErr)
			}

			if downloadErr != nil {
				fmt.Printf("download failed for %s: %v\\n", link, downloadErr)
				return downloadErr
			}
			results[index] = fmt.Sprintf("File %s processed successfully", link)
			return nil

		})

	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return results, nil

}

// actual HTTP request to get the file bytes
func downloadHelper(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w/n", err)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("failed to request: %w/n", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)

}
