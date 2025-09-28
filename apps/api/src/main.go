// apps/api-gateway/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/omar-gamzatov/content-guardian/apps/api/src/server"
)

func main() {
	server := server.NewServer()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	log.Printf("api-gateway listening on :%s", port)
	if err := http.ListenAndServe(":"+port, server.Mux); err != nil {
		log.Fatal(err)
	}
}
