package main

import (
	"log"

	"async-file-storage/internal/repository"
	"async-file-storage/internal/temporal"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	dsn := "postgres://galym:galik2006@localhost:5433/downloader?sslmode=disable"
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
