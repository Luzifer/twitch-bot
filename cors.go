package main

import (
	"net/http"
	"strings"
)

func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow the client to send us credentials
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// We only care about certain headers, so whitelist them
		w.Header().Set("Access-Control-Allow-Headers", strings.Join([]string{
			"Accept",
			"Authorization",
			"Content-Type",
			"User-Agent",
		}, ", "))

		// List all accepted methods no matter whether they are accepted by the specified endpoint
		w.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{
			http.MethodDelete,
			http.MethodGet,
			http.MethodPatch,
			http.MethodPost,
			http.MethodPut,
		}, ", "))

		// Public API: Let everyone in
		if origin := r.Header.Get("Origin"); origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		h.ServeHTTP(w, r)
	})
}
