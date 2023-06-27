package main

import (
	"context"
	"log"
)

func main() {

	store, err := NewPostgresStorage()
	defer store.db.Close(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Println("as", err)
		log.Fatal(err)
	}

	server := NewApiServer(":8080", store)
	server.Run()
}
