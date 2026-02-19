package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"async-file-storage/internal/repository"
	"async-file-storage/internal/temporal"

	"github.com/joho/godotenv"
	"go.temporal.io/sdk/client"
)

func main() {
	// TODO: структура Config
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
