package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

// Models
type Model struct {
	ID      int    `json:"id"`
	FileURL string `json:"file_url"`
}

const (
	serverName = "models-service"
)

func main() {
	// Set up HTTP routes
	http.HandleFunc("/models/{id}", handleModels)
	http.HandleFunc("/health", handleHealth)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting models microservice on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Server-Name", serverName)
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

// Model handlers
func handleModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Server-Name", serverName)

	switch r.Method {
	case "GET":
		idString := r.PathValue("id")
		if idString != "" {
			ID, err := strconv.Atoi(idString)
			if err != nil {
				http.Error(w, "Bad ID", http.StatusBadRequest)
			} else {
				getModelByID(w, ID)
			}
		} else {
			http.Error(w, "ID mustn't be empty", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getModelByID(w http.ResponseWriter, ID int) {
	m := Model{
		ID:      ID,
		FileURL: fmt.Sprintf("https://models.alexandrite.ru/%d", ID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}
