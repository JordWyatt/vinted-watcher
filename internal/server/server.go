package server

import (
	"fmt"
	"net/http"
	"vinted-watcher/internal/storage"
)

// POST /api/searches    # Add search from URL
// GET  /api/searches    # List all searches
// DELETE /api/searches/:id
// GET  /api/health
type Server interface {
	Start() error
	CreateAlertHandler(w http.ResponseWriter, r *http.Request)
	ListAlertsHandler(w http.ResponseWriter, r *http.Request)
}

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
