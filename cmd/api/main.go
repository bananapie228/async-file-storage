package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	temporaladapter "async-file-storage/internal/adapters/temporal"
	"async-file-storage/internal/repository"
	httptransport "async-file-storage/internal/transport/http"
	"async-file-storage/internal/usecase"

	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"
)

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dsn = "postgres://" + user + ":" + password + "@localhost:5433/downloader?sslmode=disable"
	}
	repo, err := repository.NewPostgresRepository(dsn)
	if err != nil {
		log.Fatalf("Failed to init repository: %v", err)
	}

	tc, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer tc.Close()

	downloader := temporaladapter.NewDownloader(tc, "file-storage-tasks")
	service := usecase.NewService(repo, downloader)
	handler := httptransport.NewHandler(service)

	var finalHandler http.Handler = handler
	finalHandler = httptransport.RequestID(finalHandler)
	finalHandler = httptransport.Recovery(finalHandler)
	finalHandler = httptransport.Logging(finalHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: finalHandler,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("API Server started on :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	// Graceful Shutdown
	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
