package main

import (
	"context"
	"log"
	"time"

	"async-file-storage/internal/repository"
	"async-file-storage/internal/temporal"

	"go.temporal.io/sdk/client"
)

func main() {
	dsn := "postgres://galym:galik2006@localhost:5433/downloader?sslmode=disable"
	repo, _ := repository.NewPostgresRepository(dsn)

	urls := []string{
		"https://raw.githubusercontent.com/temporalio/samples-go/master/README.md",
		"https://google.com",
	}
	timeout := 60 * time.Second

	requestID, err := repo.CreateRequest(context.Background(), urls)
	if err != nil {
		log.Fatalf("Failed to create request in DB: %v", err)
	}

	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	// 4. Запускаем Workflow
	options := client.StartWorkflowOptions{
		ID:        "download_request_test",
		TaskQueue: "file-storage-tasks",
	}

	we, err := c.ExecuteWorkflow(context.Background(), options, temporal.DownloadWorkflow, requestID, urls, timeout)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Printf("Workflow started! RequestID: %d, RunID: %s\n", requestID, we.GetRunID())
}
