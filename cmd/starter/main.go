package main

import (
	"context"
	"fmt"
	"log"

	"go.temporal.io/sdk/client"

	t "async-file-storage/internal/temporal"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("connection error", err)
	}
	defer c.Close()

	// 2. Запускаем Workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        "test-workflow-id",
		TaskQueue: "test-queue",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, t.SayHelloWorkflow, "gala")
	if err != nil {
		log.Fatalln("could not load workflow", err)
	}

	fmt.Printf("\n Workflow running! ID: %s, RunID: %s\n", we.GetID(), we.GetRunID())

}
