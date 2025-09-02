package server

import (
	"net/http"
)

func (s *HTTPServer) RunScraper(r *http.Request, w http.ResponseWriter) {
	_, err := s.Scraper.Scrape()
	if err != nil {
		http.Error(w, "Failed to run scraper", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
