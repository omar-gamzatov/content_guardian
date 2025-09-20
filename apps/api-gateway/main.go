// apps/api-gateway/main.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Заглушка под будущий endpoint модерации текста
	mux.HandleFunc("/v1/moderate/text", func(w http.ResponseWriter, r *http.Request) {
		type Req struct {
			TenantID string `json:"tenant_id"`
			Content  string `json:"content"`
		}
		var req Req
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		resp := map[string]any{
			"status":  "not_implemented",
			"message": "text moderation will be wired in next steps",
			"time":    time.Now().UTC(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}
	log.Printf("api-gateway listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
