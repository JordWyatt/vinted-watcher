package server

import (
	"fmt"
	"net/http"
	"vinted-watcher/internal/storage"
)

type Server interface {
	Start() error
	CreateSearchHandler(w http.ResponseWriter, r *http.Request)
	ListSearchesHandler(w http.ResponseWriter, r *http.Request)
}

// TODO: Unit Test
type HTTPServer struct {
	Storage *storage.DB
}

func NewServer(storage *storage.DB) *HTTPServer {
	return &HTTPServer{
		Storage: storage,
	}
}

func (s *HTTPServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /searches", s.CreateSearchHandler)
	mux.HandleFunc("GET /searches", s.ListSearchesHandler)

	fmt.Println("Starting server on :8080")
	return http.ListenAndServe(":8080", mux)
}
