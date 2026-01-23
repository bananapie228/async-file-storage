package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/jackc/pgx/v5"
)

func main() {
	_ = godotenv.Load()
	url := os.Getenv("DB_URL")

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	err = conn.Ping(context.Background())
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("success")
}
