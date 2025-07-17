package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	uttt "uttt/internal/engine"
	"uttt/internal/server"
	"uttt/internal/utils"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

var workerPool *server.WorkerPool

func main() {
	// Initialize the ultimate tic tac toe lib
	uttt.Init()

	// Load .env file (if exists)
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Failed to load .env file, using the default settings.")
	}

	// Set the parameters for worker pool
	workers := utils.GetEnvInt("WORKERS", 4)
	queueSize := utils.GetEnvInt("QUEUE_SIZE", 100)
	port := utils.GetEnv("PORT", "8080")

	log.Printf("Starting the worker pool: workers=%d, queueSize=%d\n", workers, queueSize)

	workerPool = server.NewWorkerPool(workers, queueSize)
	workerPool.Start()

	// Create new router
	server.SetupCors()
	router := mux.NewRouter()
	router.Use(server.CorsMiddleware)
	router.HandleFunc("/analysis", handleAnalysis)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Graceful shutdown, closes all remaning jobs on Ctrl+C
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		workerPool.Stop()
		_ = srv.Shutdown(workerPool.Context())
	}()

	log.Printf("Listening on: http://localhost:%s\n", port)
	log.Fatal(srv.ListenAndServe())
}

func handleAnalysis(w http.ResponseWriter, r *http.Request) {
	var req server.AnalysisRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	req.Response = make(chan server.AnalysisResponse, 1)

	// Try submitting the request
	if !workerPool.Submit(req) {
		http.Error(w, "Server is too busy", http.StatusServiceUnavailable)
		return
	}

	// Now wait for the analysis
	timeout := time.NewTimer(30 * time.Second)
	select {
	case resp := <-req.Response:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	case <-timeout.C:
		http.Error(w, "Analysis timeout", http.StatusRequestTimeout)
	}
}
