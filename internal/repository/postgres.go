package repository

import (
	"async-file-storage/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(dsn string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", dsn)
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS requests (
        id SERIAL PRIMARY KEY,
        status TEXT NOT NULL,
        created_at TIMESTAMP NOT NULL
    );
`)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
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
		return nil, fmt.Errorf("failed to create files table: %w", err)
	}
	return &PostgresRepository{db: db}, nil
}
func (r *PostgresRepository) CreateRequest(ctx context.Context, urls []string) (int, error) {
	// begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// create request with status "Process"

	var requestID int
	err = tx.QueryRowContext(ctx,
		"INSERT INTO requests(status, created_at) VALUES ($1, $2) RETURNING id",
		domain.StatusProcess, time.Now(),
	).Scan(&requestID)

	//return error if failed to insert

	if err != nil {
		return 0, fmt.Errorf("failed to insert: %w", err)
	}

	query := "INSERT INTO files (request_ID,  url) VALUES ($1, $2)"
	for _, url := range urls {
		_, err = tx.ExecContext(ctx, query, requestID, url)
		if err != nil {
			return 0, fmt.Errorf("failed to insert files: %w", err)
		}
	}
	// commit transaction
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return requestID, nil
}

// change status of request
func (r *PostgresRepository) UpdateRequestStatus(ctx context.Context, id int, status domain.Status) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE requests SET status = $1 WHERE id = $2",
		status, id)
	return err
}

// save files or errors
func (r *PostgresRepository) UpdateFileStatus(ctx context.Context, requestID int, url string, data []byte, downloadErr error) error {
	var errMsg string
	if downloadErr != nil {
		errMsg = downloadErr.Error()
	}

	_, err := r.db.ExecContext(ctx,
		"UPDATE files SET data = $1, err_Msg = $2 WHERE requetID = $3 AND url = %4",
		data, errMsg, requestID, url)
	return err
}

func (r *PostgresRepository) GetRequestStatus(ctx context.Context, id int) (*domain.DownloadRequest, []domain.FileEntry, error) {
	req := &domain.DownloadRequest{}
	err := r.db.QueryRowContext(ctx,
		"SELECT id, status, created_at FROM requests WHEERE id = $1", id,
	).Scan(&req.ID, &req.Status, &req.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil, fmt.Errorf("request not found")
	}
	if err != nil {
		return nil, nil, err
	}
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, request_id, url, error_msg FROM files WHERE request_id = $1", id,
	)
	defer rows.Close()

	if err != nil {
		return nil, nil, err
	}
	var files []domain.FileEntry
	for rows.Next() {
		var f domain.FileEntry
		// error_msg might be NULL, use sql.NullString for scanning
		var dbErr sql.NullString

		if err := rows.Scan(&f.ID, &f.RequestID, &f.URL, &dbErr); err != nil {
			return nil, nil, err
		}
		f.Error = dbErr.String // Превращаем NULL в пустую строку
		files = append(files, f)
	}

	return req, files, nil
}
