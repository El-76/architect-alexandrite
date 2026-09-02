package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"go.opentelemetry.io/otel/trace"
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
	ctx := context.Background()

	tracer := initTracer(ctx, serverName)

	// Set up HTTP routes
	http.HandleFunc("/orders/{id}", func(w http.ResponseWriter, r *http.Request) { handleOrders(w, r, tracer) })
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { handleHealth(w, r, tracer) })

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting orders microservice on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request, tracer trace.Tracer) {
	_, span := tracer.Start(r.Context(), "/health")
	defer span.End()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Server-Name", serverName)
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

// Order handlers
func handleOrders(w http.ResponseWriter, r *http.Request, tracer trace.Tracer) {
	ctx, span := tracer.Start(r.Context(), "/orders")
	defer span.End()

	w.Header().Set("X-Server-Name", serverName)

	switch r.Method {
	case "GET":
		idString := r.PathValue("id")
		if idString != "" {
			ID, err := strconv.Atoi(idString)
			if err != nil {
				http.Error(w, "Bad ID", http.StatusBadRequest)
			} else {
				err = getOrderByID(ctx, w, ID)

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

func getOrderByID(ctx context.Context, w http.ResponseWriter, ID int) error {
	modelsServiceBaseURL := os.Getenv("MODELS_SERVICE_BASE_URL")
	if modelsServiceBaseURL == "" {
		modelsServiceBaseURL = "http://models-service:8080"
	}

	o := Order{
		ID:      ID,
		ModelID: ID + 10000,
		Status:  "FILE_UPLOADED",
	}

	m, err := getModelByID(ctx, modelsServiceBaseURL, o.ModelID)

	if err != nil {
		return err
	}

	o.ModelFileURL = m.FileURL

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(o)

	return nil
}

func getModelByID(ctx context.Context, baseURL string, modelID int) (*Model, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("%s/models/%d", baseURL, modelID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed GET request to %s: %w", url, err)
	}

	propagator := propagation.TraceContext{}

	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	resp, err := client.Do(req)
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

func initTracer(ctx context.Context, service string) trace.Tracer {
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
	)
	exporter, err := otlptrace.New(ctx, client)
	if err != nil {
		log.Fatal("creating OTLP trace exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(service),
		)),
	)

	return tp.Tracer(service)
}
