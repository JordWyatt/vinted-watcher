package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// TODO: Unit Test
func (s *HTTPServer) ListSearchesHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Listing all searches")
	searches, err := s.Storage.GetAllSearches()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	slog.Info("Successfully retrieved all searches", "count", len(searches))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searches)
}
