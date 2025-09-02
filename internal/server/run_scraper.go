package server

import (
	"net/http"
)

func (s *HTTPServer) RunScraperHandler(w http.ResponseWriter, r *http.Request) {
	_, err := s.Scraper.Scrape()
	if err != nil {
		http.Error(w, "Failed to run scraper", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
