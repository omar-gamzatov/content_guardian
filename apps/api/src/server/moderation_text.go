package server

import (
	"encoding/json"
	"net/http"
	"time"
)

func moderateText(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		TenantID string `json:"tenant_id"`
		Content  string `json:"content"`
	}
	var req Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	resp := map[string]string{
		"status":  "not_implemented",
		"message": "text moderation will be wired in next steps",
		"time":    time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
