package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
	"vinted-watcher/internal/storage"
)

type Server interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
	CreateSearchHandler(w http.ResponseWriter, r *http.Request)
	ListSearchesHandler(w http.ResponseWriter, r *http.Request)
}

// TODO: Unit Test
type HTTPServer struct {
	Storage    *storage.DB
	httpServer *http.Server
}

func NewServer(storage *storage.DB) *HTTPServer {
	return &HTTPServer{
		Storage: storage,
	}
}

func (s *HTTPServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /searches", s.CreateSearchHandler)
	mux.HandleFunc("GET /searches", s.ListSearchesHandler)

	s.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	slog.Info("Starting server on :8080")

	errChan := make(chan error)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	slog.Info("Server listening on :8080")

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		slog.Info("Context cancelled. Shutting down server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	}
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	slog.Info("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}
