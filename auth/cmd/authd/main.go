package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const apiKey = "dummy-api-key"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Info("Authorization header: %s\n", r.Header.Get("Authorization"))
		if r.Header.Get("Authorization") == apiKey {
			w.WriteHeader(http.StatusOK)
			if _, err := fmt.Fprint(w, "Authenticated"); err != nil {
				return
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			if _, err := fmt.Fprint(w, "Unauthorized"); err != nil {
				return
			}
		}
	})

	log.Info("Starting auth daemon on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Auth daemon failed to start: %v", err)
	}
}
