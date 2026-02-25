package repository

import (
	"async-file-storage/internal/domain"
	"context"
	"database/sql"
	"errors" // Добавили для проверки ошибок
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

// creates a new instance of the repository.
func NewPostgresRepository(dsn string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS requests (
        id SERIAL PRIMARY KEY,
        status TEXT NOT NULL,
        created_at TIMESTAMP NOT NULL
    );
`)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create requests table: %w", err)
	}

	_, err = db.Exec(`
       CREATE TABLE IF NOT EXISTS files (
          id SERIAL PRIMARY KEY,
          request_id INTEGER REFERENCES requests(id),
          url TEXT NOT NULL,
          data BYTEA, 
          error_msg TEXT
       );
    `)

	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create files table: %w", err)
	}
	return &PostgresRepository{db: db}, nil
}

// creates a new download request and its file entries.
func (r *PostgresRepository) CreateRequest(ctx context.Context, urls []string) (int, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		// TODO: лучше commit и rollback обрабатывать здесь, если ошибки нет, то вызовется commit, иначе rollback
		_ = tx.Rollback()
	}()

	var requestID int
	err = tx.QueryRowContext(ctx,
		"INSERT INTO requests(status, created_at) VALUES ($1, $2) RETURNING id",
		domain.StatusProcess, time.Now(),
	).Scan(&requestID)

	if err != nil {
		return 0, fmt.Errorf("failed to insert: %w", err)
	}

	query := "INSERT INTO files (request_id, url) VALUES ($1, $2)"
	for _, url := range urls {
		_, err = tx.ExecContext(ctx, query, requestID, url)
		if err != nil {
			return 0, fmt.Errorf("failed to insert files: %w", err)
		}
	}

	// можно перенести commit в defer
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return requestID, nil
}

// UpdateRequestStatus changes the status of a specific request.
func (r *PostgresRepository) UpdateRequestStatus(ctx context.Context, id int, status domain.Status) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE requests SET status = $1 WHERE id = $2",
		status, id)
	return err
}

// UpdateFileStatus saves file data or records an error message.
func (r *PostgresRepository) UpdateFileStatus(ctx context.Context, requestID int, url string, data []byte, downloadErr error) error {
	var errMsg string
	if downloadErr != nil {
		errMsg = downloadErr.Error()
	}

	_, err := r.db.ExecContext(ctx,
		"UPDATE files SET data = $1, error_msg = $2 WHERE request_id = $3 AND url = $4",
		data, errMsg, requestID, url)
	return err
}

// GetRequestStatus returns the request and all associated files.
// TODO: как думаешь хорошо ли что мы возвращаем fileEntry где есть поле data, которое может занимать много памяти,
// может стоит возвращать его только в GetFile, а здесь возвращать только метаинформацию о файлах, например id, url и error?
// Представь что у тебя есть файл с размером 1ГБ, и ты хочешь получить статус запроса, тебе не нужно загружать весь этот файл в память,
// а так как у тебя сейчас устроено, ты его загрузишь, а потом просто не будешь использовать, что может привести к проблемам с памятью
func (r *PostgresRepository) GetRequestStatus(ctx context.Context, id int) (*domain.DownloadRequest, []domain.FileEntry, error) {
	req := &domain.DownloadRequest{}
	// Поправил WHEERE -> WHERE
	err := r.db.QueryRowContext(ctx,
		"SELECT id, status, created_at FROM requests WHERE id = $1", id,
	).Scan(&req.ID, &req.Status, &req.CreatedAt)

	// TODO: сначала лучше сделать if err != nil, а внутри него уже проверять на sql.ErrNoRows и на др. ошибку
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}

	rows, err := r.db.QueryContext(ctx,
		"SELECT id, request_id, url, error_msg FROM files WHERE request_id = $1", id,
	)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = rows.Close() }()

	var files []domain.FileEntry
	for rows.Next() {
		var f domain.FileEntry
		var dbErr sql.NullString

		if err := rows.Scan(&f.ID, &f.RequestID, &f.URL, &dbErr); err != nil {
			return nil, nil, err
		}
		f.Error = dbErr.String
		files = append(files, f)
	}

	return req, files, nil
}

// GetFile returns a file by request and file id.
func (r *PostgresRepository) GetFile(ctx context.Context, requestID int, fileID int) (*domain.FileEntry, error) {
	var f domain.FileEntry
	var dbErr sql.NullString

	err := r.db.QueryRowContext(ctx,
		"SELECT id, request_id, url, data, error_msg FROM files WHERE request_id = $1 AND id = $2",
		requestID, fileID,
	).Scan(&f.ID, &f.RequestID, &f.URL, &f.Data, &dbErr)

	// TODO: сначала лучше сделать if err != nil, а внутри него уже проверять на sql.ErrNoRows и на др. ошибку
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	f.Error = dbErr.String
	return &f, nil
}
