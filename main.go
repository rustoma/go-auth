package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

type application struct {
	DB *pgx.Conn
}

func main() {
	var app application

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	log.Println("Starting application on port", os.Getenv("PORT"))

	//start a web server
	err = http.ListenAndServe(fmt.Sprintf(":%s", os.Getenv("PORT")), app.routes())
	if err != nil {
		log.Fatal(err)
	}

}
