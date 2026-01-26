package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	t "async-file-storage/internal/temporal"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("could not connect Temporal Client", err)
	}
	defer c.Close()

	w := worker.New(c, "test-queue", worker.Options{})

	w.RegisterWorkflow(t.SayHelloWorkflow)
	w.RegisterActivity(t.Greet)

	// 4. Запускаем
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("worker failed", err)
	}
}
