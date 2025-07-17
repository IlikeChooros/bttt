package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	uttt "uttt/internal/engine"
	"uttt/internal/server"

	"github.com/gorilla/mux"
)

var workerPool *server.WorkerPool

func main() {
	// Initialize the ultimate tic tac toe lib
	uttt.Init()
	server.LoadConfig()

	// Set the parameters for worker pool
	log.Printf("Starting the worker pool: workers=%d, queueSize=%d\n",
		server.DefaultConfig.Pool.DefaultWorkers,
		server.DefaultConfig.Pool.DefaultQueueSize)

	ctx := context.Background()
	workerPool = server.NewWorkerPool(
		server.DefaultConfig.Pool.DefaultWorkers,
		server.DefaultConfig.Pool.DefaultQueueSize)
	workerPool.Start(ctx)

	// Create new logger
	logger := server.NewLogger()

	// Create new router
	router := mux.NewRouter()
	router.Use(server.TracingMiddleware())
	router.Use(server.LoggingMiddleware(logger))
	router.Use(server.MetricsMiddleware)
	router.Use(server.CorsMiddleware)

	// API endpoints
	router.HandleFunc("/analysis", server.AnalysisHandler(workerPool))
	router.HandleFunc("/health", server.HealthHandler(workerPool))
	router.HandleFunc("/metrics", server.MetricsHandler())

	srv := &http.Server{
		Addr:         ":" + server.DefaultConfig.Server.Port,
		Handler:      router,
		ReadTimeout:  server.DefaultConfig.Server.ReadTimeout,
		WriteTimeout: server.DefaultConfig.Server.WriteTimeout,
	}

	// Graceful shutdown, closes all remaning jobs on Ctrl+C
	done := make(chan bool, 1)
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		workerPool.Stop()
		ctx, cancel := context.WithTimeout(ctx, server.DefaultConfig.Server.ShutdownTimeout)

		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			log.Println("Couldn't shutdown server gracefully...")
		}
		close(done)
	}()

	log.Printf("Listening on: http://localhost:%s\n", server.DefaultConfig.Server.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	<-done
	log.Println("Server stopped")
}
