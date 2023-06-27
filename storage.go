package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

type Storage interface {
	InsertUser(user User) (int, error)
}

type PostgresStorage struct {
	db      *pgx.Conn
	timeout time.Duration
}

func NewPostgresStorage() (*PostgresStorage, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		if err != nil {
			return nil, err
		}
		os.Exit(1)
	}

	return &PostgresStorage{
		db:      conn,
		timeout: time.Second * 3,
	}, nil
}

func (s *PostgresStorage) Init() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	createUserTableStmt := `CREATE TABLE IF NOT EXISTS public.user (
							id SERIAL PRIMARY KEY,
							user_name VARCHAR(255),
							password VARCHAR(255),
							created_at TIMESTAMPTZ,
							updated_at TIMESTAMPTZ
							)`

	_, err := s.db.Exec(ctx, createUserTableStmt)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) InsertUser(user User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	createUserStmt := `insert into public.user (user_name,password,created_at,updated_at)
			values ($1,$2,$3,$4) returning id`

	var userID int

	err := s.db.QueryRow(ctx, createUserStmt,
		user.UserName,
		user.Password,
		time.Now().UTC(),
		time.Now().UTC(),
	).Scan(&userID)

	return userID, err

}
