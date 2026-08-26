package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
)

// Models
type Order struct {
	ID           int    `json:"id"`
	Status       string `json:"status"`
	ModelID      int    `json:"model_id"`
	ModelFileURL string `json:"model_file_url"`
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
				getOrderByID(w, ID)
			}
		} else {
			http.Error(w, "ID mustn't be empty", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getOrderByID(w http.ResponseWriter, ID int) {
	o := Order{
		ID:      ID,
		ModelID: ID + 10000,
		Status:  "FILE_UPLOADED",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)
}
