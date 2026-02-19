package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"async-file-storage/internal/domain"
)

type Service struct {
	repo       Repository
	downloader Downloader
}

func NewService(repo Repository, downloader Downloader) *Service {
	return &Service{repo: repo, downloader: downloader}
}

func (s *Service) CreateRequest(ctx context.Context, input CreateRequestInput) (CreateRequestOutput, error) {
	if len(input.URLs) == 0 || input.Timeout <= 0 {
		return CreateRequestOutput{}, ErrInvalidInput
	}
	for _, url := range input.URLs {
		if strings.TrimSpace(url) == "" {
			return CreateRequestOutput{}, ErrInvalidInput
		}
	}

	requestID, err := s.repo.CreateRequest(ctx, input.URLs)
	if err != nil {
		return CreateRequestOutput{}, fmt.Errorf("create request: %w", err)
	}

	if err := s.downloader.StartDownload(ctx, requestID, input.URLs, input.Timeout); err != nil {
		return CreateRequestOutput{}, fmt.Errorf("start download: %w", err)
	}

	return CreateRequestOutput{ID: requestID, Status: domain.StatusProcess}, nil
}

func (s *Service) GetRequest(ctx context.Context, id int) (GetRequestOutput, error) {
	if id <= 0 {
		return GetRequestOutput{}, ErrInvalidInput
	}

	req, files, err := s.repo.GetRequestStatus(ctx, id)
	if err != nil {
		// TODO: обрабатываешь одну и ту же ошибку только из разных пакетов,
		// может стоит унифицировать ошибку NotFound в одном пакете и использовать её везде?
		if errors.Is(err, ErrNotFound) || errors.Is(err, domain.ErrNotFound) {
			return GetRequestOutput{}, ErrNotFound
		}
		return GetRequestOutput{}, fmt.Errorf("get request: %w", err)
	}

	out := GetRequestOutput{ID: req.ID, Status: req.Status}
	out.Files = make([]FileStatus, 0, len(files))
	for _, f := range files {
		status := FileStatus{URL: f.URL}
		// лучше возвращать указатель чтобы сравнивать через f.Error != nil, а не через пустую строку,
		// так как может быть ситуация когда ошибка есть, но она не описана, и тогда будет возвращаться пустая строка,
		// что может ввести в заблуждение
		if f.Error != "" {
			status.ErrorCode = f.Error
		} else {
			status.FileID = f.ID
		}
		out.Files = append(out.Files, status)
	}
	return out, nil
}

func (s *Service) GetFile(ctx context.Context, requestID int, fileID int) (GetFileOutput, error) {
	if requestID <= 0 || fileID <= 0 {
		return GetFileOutput{}, ErrInvalidInput
	}

	file, err := s.repo.GetFile(ctx, requestID, fileID)
	if err != nil {
		// TODO: та же ситуация с ошибкой NotFound, надо унифицировать её в одном пакете и использовать везде,
		// чтобы не проверять её из разных пакетов
		if errors.Is(err, ErrNotFound) || errors.Is(err, domain.ErrNotFound) {
			return GetFileOutput{}, ErrNotFound
		}
		return GetFileOutput{}, fmt.Errorf("get file: %w", err)
	}
	if file.Error != "" {
		return GetFileOutput{}, BusinessError{Code: file.Error, Msg: "file not available"}
	}

	return GetFileOutput{Data: file.Data}, nil
}
