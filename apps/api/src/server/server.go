package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type RulesEngineClient struct {
	HTTPClient *http.Client
	BaseURL    string
}

func NewRulesEngineClient(baseURL string) *RulesEngineClient {
	return &RulesEngineClient{
		HTTPClient: &http.Client{
			Timeout: time.Second * 5, // Don't forget about timeouts
		},
		BaseURL: baseURL,
	}
}

func (c *RulesEngineClient) ModerateText(content string) (string, error) {
	body, _ := json.Marshal(map[string]string{"content": content})
	resp, err := c.HTTPClient.Post(c.BaseURL+"/moderate/text", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data struct {
		Result string `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	return data.Result, nil
}

type Server struct {
	Mux *http.ServeMux
}

func NewServer() *Server {
	server := &Server{
		Mux: http.NewServeMux(),
	}
	server.Mux.HandleFunc("/health", healthCheck)

	// Заглушка под будущий endpoint модерации текста
	server.Mux.HandleFunc("/v1/moderate/text", moderateText)
	return server
}
