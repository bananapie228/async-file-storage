package main

import (
	"fmt"
	"log"
	"os"

	"async-file-storage/internal/repository"
	"async-file-storage/internal/temporal"

	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		host := os.Getenv("DB_HOST")
		if host == "" {
			host = "localhost"
		}
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5433"
		}
		name := os.Getenv("DB_NAME")
		if name == "" {
			name = "downloader"
		}
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, name)
	}
	repo, err := repository.NewPostgresRepository(dsn)
	if err != nil {
		log.Fatalf("Failed to init repository: %v", err)
	}

	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer c.Close()

	w := worker.New(c, "file-storage-tasks", worker.Options{})

	w.RegisterWorkflow(temporal.DownloadWorkflow)

	activityContainer := &temporal.Activities{Repo: repo}
	w.RegisterActivity(activityContainer)

	log.Println("Worker is starting...")
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Worker failed: %v", err)
	}
}
