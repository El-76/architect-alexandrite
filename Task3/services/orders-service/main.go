package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Models
type Order struct {
	ID           int    `json:"id"`
	Status       string `json:"status"`
	ModelID      int    `json:"model_id"`
	ModelFileURL string `json:"model_file_url"`
}

type Model struct {
	ID      int    `json:"id"`
	FileURL string `json:"file_url"`
}

const (
	serverName = "orders-service"
)

func main() {
	// Set up HTTP routes
	http.HandleFunc("/orders/{id}", handleOrders)
	http.HandleFunc("/health", handleHealth)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting orders microservice on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Server-Name", serverName)
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

// Order handlers
func handleOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Server-Name", serverName)

	switch r.Method {
	case "GET":
		idString := r.PathValue("id")
		if idString != "" {
			ID, err := strconv.Atoi(idString)
			if err != nil {
				http.Error(w, "Bad ID", http.StatusBadRequest)
			} else {
				err = getOrderByID(w, ID)

				if err != nil {
					http.Error(w, "Internal error", http.StatusInternalServerError)
				}
			}
		} else {
			http.Error(w, "ID mustn't be empty", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getOrderByID(w http.ResponseWriter, ID int) error {
	modelsServiceBaseURL := os.Getenv("MODELS_SERVICE_BASE_URL")
	if modelsServiceBaseURL == "" {
		modelsServiceBaseURL = "http://models-service:8080"
	}

	o := Order{
		ID:      ID,
		ModelID: ID + 10000,
		Status:  "FILE_UPLOADED",
	}

	m, err := getModelByID(modelsServiceBaseURL, o.ModelID)

	if err != nil {
		return err
	}

	o.ModelFileURL = m.FileURL

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)

	return nil
}

func getModelByID(baseURL string, modelID int) (*Model, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/models/%d", baseURL, modelID)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching model data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var model Model
	if err := json.NewDecoder(resp.Body).Decode(&model); err != nil {
		return nil, fmt.Errorf("error decoding telemetry response: %w", err)
	}

	return &model, nil
}
