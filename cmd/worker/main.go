package main

import (
	"log"

	"async-file-storage/internal/config"
	"async-file-storage/internal/repository"
	"async-file-storage/internal/temporal"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	cfg := config.Load()

	repo, err := repository.NewPostgresRepository(cfg.DBURL)
	if err != nil {
		log.Fatalf("Failed to init repository: %v", err)
	}

	c, err := client.Dial(client.Options{
		HostPort: cfg.TemporalHost,
	})
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
