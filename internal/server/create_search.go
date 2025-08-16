package server

import (
	"encoding/json"
	"net/http"
	"vinted-watcher/internal/domain"
	"vinted-watcher/internal/vinted"
)

type CreateAlertRequest struct {
	URL string `json:"url"`
}

type CreateAlertResponse struct {
	ID int `json:"id"`
}

// TODO: Unit Test
func (s *HTTPServer) CreateSearchHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "Missing URL", http.StatusBadRequest)
		return
	}

	searchParams, err := vinted.ParseVintedURL(req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	savedSearch := domain.NewSavedSearch(searchParams)

	searchID, err := s.Storage.CreateSearch(savedSearch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := CreateAlertResponse{
		ID: searchID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
