package server

import (
	"encoding/json"
	"net/http"
)

// TODO: Unit Test
func (s *HTTPServer) ListSearchesHandler(w http.ResponseWriter, r *http.Request) {
	searches, err := s.Storage.GetAllSearches()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(searches)
}
