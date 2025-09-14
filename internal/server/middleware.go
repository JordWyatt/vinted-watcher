package server

import (
	"net/http"
	"os"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bearer token authentication
		expectedToken := os.Getenv("API_TOKEN")
		expectedAuthorizationHeader := "Bearer " + expectedToken

		if r.Header.Get("Authorization") != expectedAuthorizationHeader {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
