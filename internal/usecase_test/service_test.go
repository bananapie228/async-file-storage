package usecase_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"async-file-storage/internal/domain"
	"async-file-storage/internal/usecase"
)

type mockRepo struct {
	createRequestFunc func(ctx context.Context, urls []string) (int, error)
	getRequestFunc    func(ctx context.Context, id int) error
	getFileFunc       func(ctx context.Context, requestID int, fileID int) error
}

func (m *mockRepo) CreateRequest(ctx context.Context, urls []string) (int, error) {
	return m.createRequestFunc(ctx, urls)
}

func (m *mockRepo) GetRequestStatus(ctx context.Context, id int) (*domain.DownloadRequest, []domain.FileEntry, error) {
	if m.getRequestFunc != nil {
		if err := m.getRequestFunc(ctx, id); err != nil {
			return nil, nil, err
		}
	}
	return &domain.DownloadRequest{ID: id}, nil, nil
}

func (m *mockRepo) GetFile(ctx context.Context, requestID int, fileID int) (*domain.FileEntry, error) {
	if m.getFileFunc != nil {
		if err := m.getFileFunc(ctx, requestID, fileID); err != nil {
			return nil, err
		}
	}
	return &domain.FileEntry{ID: fileID, RequestID: requestID}, nil
}

type mockDownloader struct {
	startFunc func(ctx context.Context, requestID int, urls []string, timeout time.Duration) error
}

func (m *mockDownloader) StartDownload(ctx context.Context, requestID int, urls []string, timeout time.Duration) error {
	return m.startFunc(ctx, requestID, urls, timeout)
}

func TestServiceCreateRequest_Success(t *testing.T) {
	repo := &mockRepo{}
	downloader := &mockDownloader{}

	expectedURLs := []string{"https://example.com/a", "https://example.com/b"}
	expectedTimeout := 30 * time.Second

	repo.createRequestFunc = func(ctx context.Context, urls []string) (int, error) {
		if !reflect.DeepEqual(urls, expectedURLs) {
			return 0, errors.New("unexpected urls")
		}
		return 42, nil
	}
	downloader.startFunc = func(ctx context.Context, requestID int, urls []string, timeout time.Duration) error {
		if requestID != 42 {
			return errors.New("unexpected request id")
		}
		if !reflect.DeepEqual(urls, expectedURLs) {
			return errors.New("unexpected urls")
		}
		if timeout != expectedTimeout {
			return errors.New("unexpected timeout")
		}
		return nil
	}

	svc := usecase.NewService(repo, downloader)
	out, err := svc.CreateRequest(context.Background(), usecase.CreateRequestInput{
		URLs:    expectedURLs,
		Timeout: expectedTimeout,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ID != 42 {
		t.Fatalf("expected id 42, got %d", out.ID)
	}
	if out.Status != domain.StatusProcess {
		t.Fatalf("expected status PROCESS, got %s", out.Status)
	}
}

func TestServiceCreateRequest_InvalidInput(t *testing.T) {
	repo := &mockRepo{createRequestFunc: func(ctx context.Context, urls []string) (int, error) {
		return 0, errors.New("should not be called")
	}}
	downloader := &mockDownloader{startFunc: func(ctx context.Context, requestID int, urls []string, timeout time.Duration) error {
		return errors.New("should not be called")
	}}

	svc := usecase.NewService(repo, downloader)
	_, err := svc.CreateRequest(context.Background(), usecase.CreateRequestInput{
		URLs:    nil,
		Timeout: 10 * time.Second,
	})
	if !errors.Is(err, usecase.ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
