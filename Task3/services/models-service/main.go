package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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
type Model struct {
	ID      int    `json:"id"`
	FileURL string `json:"file_url"`
}

const (
	serverName = "models-service"
)

func main() {
	ctx := context.Background()

	tracer := initTracer(ctx, serverName)

	// Set up HTTP routes
	http.HandleFunc("/models/{id}", func(w http.ResponseWriter, r *http.Request) { handleModels(w, r, tracer) })
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { handleHealth(w, r, tracer) })

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting models microservice on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request, tracer trace.Tracer) {
	_, span := tracer.Start(r.Context(), "/health")
	defer span.End()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Server-Name", serverName)
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

// Model handlers
func handleModels(w http.ResponseWriter, r *http.Request, tracer trace.Tracer) {
	p := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	ctx := p.Extract(r.Context(), propagation.HeaderCarrier(r.Header))

	_, span := tracer.Start(ctx, "/models")
	defer span.End()

	if rand.Intn(2) == 0 {
		time.Sleep(1 * time.Second)
	}

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
