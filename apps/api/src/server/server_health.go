package server

import (
	"log"
	"net/http"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
	log.Printf("health check")
}
