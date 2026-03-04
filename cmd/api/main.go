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
	"async-file-storage/internal/config"
	"async-file-storage/internal/repository"
	httptransport "async-file-storage/internal/transport/http"
	"async-file-storage/internal/usecase"

	"go.temporal.io/sdk/client"
)

func main() {
	cfg := config.Load()

	repo, err := repository.NewPostgresRepository(cfg.DBURL)
	if err != nil {
		log.Fatalf("Failed to init repository: %v", err)
	}

	tc, err := client.Dial(client.Options{
		HostPort: cfg.TemporalHost,
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer tc.Close()

	adapter := temporaladapter.NewDownloader(tc, "file-storage-tasks")

	uc := usecase.NewService(repo, adapter)

	handler := httptransport.NewHandler(uc)

	srv := &http.Server{
		Addr:    cfg.Port,
		Handler: handler,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("API Server started on %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Listen error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down API server...")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
}
