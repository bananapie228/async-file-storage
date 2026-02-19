package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"github.com/jackc/pgx/v5"
)

// TODO: в чем смысл функции? Она просто пингует БД. Тем более у тебя это есть в репозитории.
// Также до этого ты использовал пакет "database/sql", а здесь перешел на "github.com/jackc/pgx/v5".
// Лучше не смешивать либы
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
