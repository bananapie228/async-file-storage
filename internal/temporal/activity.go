package temporal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"async-file-storage/internal/domain"
)

type Activities struct {
	// можно сделать переменную приватной, я не увидел где ты ее присваиваешь извне,
	// а так она может быть изменена в любой момент и это может привести к проблемам,
	// если кто-то случайно присвоит ей другое значение
	Repo domain.Storage
}

// download multiple files and save to the DB
// TODO: этот метод надо отрефакторить, слишком много кода
func (a *Activities) DownloadFilesActivity(ctx context.Context, requestID int, urls []string, timeout time.Duration) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var (
		wg      sync.WaitGroup
		sem     = make(chan struct{}, 3)
		results = make([]string, len(urls))
		mu      sync.Mutex
		done    = make([]bool, len(urls))
	)

	for i, url := range urls {
		if ctx.Err() != nil {
			break
		}

		wg.Add(1)
		sem <- struct{}{}

		index := i
		link := url

		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			if ctx.Err() != nil {
				return
			}

			// TODO: надо еще обрабатывать случай, если произойдет таймаут через select и ctx.Done(), чтобы не запускать загрузку, которая уже не нужна

			fmt.Printf("[%d] downloading: %s\n", index, link)

			data, downloadErr := downloadHelper(ctx, link)
			if downloadErr != nil {
				downloadErr = mapDownloadError(downloadErr)
			}

			if dbErr := a.Repo.UpdateFileStatus(ctx, requestID, link, data, downloadErr); dbErr != nil {
				fmt.Printf("db update error: %v\n", dbErr)
				return
			}

			if downloadErr == nil {
				results[index] = fmt.Sprintf("File %s processed successfully", link)
			}

			mu.Lock()
			// TODO: лучше писать defer mu.Unlock() сразу после Lock
			// если я здесь пропишу panic("something"), у тебя произойдет deadlock, потому что Unlock не вызовется,
			// а так, даже если будет паника, Unlock все равно вызовется и другие горутины не будут висеть в ожидании разблокировки
			done[index] = true
			mu.Unlock() // это в defer
		}()
	}

	wg.Wait()

	if ctx.Err() != nil {
		statusCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		for i, url := range urls {
			mu.Lock()
			alreadyDone := done[i]
			mu.Unlock()
			if alreadyDone {
				continue
			}
			_ = a.Repo.UpdateFileStatus(statusCtx, requestID, url, nil, errors.New("TIMEOUT"))
		}
	}

	statusCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.Repo.UpdateRequestStatus(statusCtx, requestID, domain.StatusDone); err != nil {
		return nil, fmt.Errorf("update request status: %w", err)
	}

	return results, nil
}

// actual HTTP request to get the file bytes
func downloadHelper(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)

}

func mapDownloadError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return errors.New("TIMEOUT")
	}
	return errors.New("DOWNLOAD_FAILED")
}
