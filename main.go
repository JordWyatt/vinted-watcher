package main

import (
	"fmt"
	"vinted-watcher/internal/server"
	"vinted-watcher/internal/storage"
)

// Test code - will eventually become server entrypoint
func main() {
	db, err := storage.NewDB("vinted.db")
	if err != nil {
		fmt.Println("Error initializing database:", err)
		return
	}

	server := server.NewServer(db)
	if err := server.Start(); err != nil {
		fmt.Println("Error starting server:", err)
	}

	err = server.Start()
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
