package server

import (
	"fmt"
	"net/http"
	"vinted-watcher/internal/storage"
)

type Server struct {
	Storage *storage.DB
}

func NewServer(storage *storage.DB) *Server {
	return &Server{
		Storage: storage,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/create-alert", s.CreateAlertHandler)

	fmt.Println("Starting server on :8080")
	return http.ListenAndServe(":8080", mux)
}
