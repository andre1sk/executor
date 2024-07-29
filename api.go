package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type APIServer struct {
	Executor Executor
	Logger   *log.Logger
	mux      *http.ServeMux
}

func NewAPIServer(executor Executor, logger *log.Logger) *APIServer {
	mux := http.NewServeMux()
	server := &APIServer{
		Executor: executor,
		Logger:   logger,
		mux:      mux,
	}
	// Register routes
	server.mux.HandleFunc("POST /enrich", server.enrichHandler)
	return server
}

func (s *APIServer) Start() error {
	s.Logger.Printf("Starting API server on http://localhost:8080")
	return http.ListenAndServe(":8080", s.mux)
}

func (s *APIServer) enrichHandler(w http.ResponseWriter, r *http.Request) {
	alerts, err := ParseAlerts(r.Body)
	if err != nil {
		s.Logger.Printf("Failed to parse alerts: %v", err)
		http.Error(w, "Failed to parse alerts", http.StatusBadRequest)
		return
	}

	enrichedAlerts, err := s.Executor.Enrich(alerts)
	if err != nil {
		s.Logger.Printf("Failed to enrich alerts: %v", err)
		http.Error(w, "Failed to enrich alerts", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(enrichedAlerts); err != nil {
		s.Logger.Printf("Failed to encode enriched alerts: %v", err)
		http.Error(w, "Failed to encode enriched alerts", http.StatusInternalServerError)
		return
	}
}

func ParseAlerts(r io.Reader) ([]Alert, error) {
	var alerts []Alert
	if err := json.NewDecoder(r).Decode(&alerts); err != nil {
		return nil, fmt.Errorf("failed to decode alerts: %v", err)
	}
	return alerts, nil
}
