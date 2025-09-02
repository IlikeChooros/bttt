package main

import (
	"context"
	"fmt"
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

	// Create new logger
	logger := server.NewLogger()

	// Set the parameters for worker pool
	logger.Info(fmt.Sprintf(
		"Starting the worker pool: workers=%d, queueSize=%d",
		server.DefaultConfig.Pool.DefaultWorkers,
		server.DefaultConfig.Pool.DefaultQueueSize,
	))

	ctx := context.Background()
	workerPool = server.NewWorkerPool(
		server.DefaultConfig.Pool.DefaultWorkers,
		server.DefaultConfig.Pool.DefaultQueueSize)
	workerPool.Start(ctx)

	// Create new router
	router := mux.NewRouter()
	router.Use(server.CorsMiddleware)
	router.Use(server.TracingMiddleware)
	router.Use(server.LoggingMiddleware(logger))
	router.Use(server.RateLimiterMiddleware(logger))
	router.Use(server.MetricsMiddleware)

	// API endpoints
	router.HandleFunc("/analysis", server.AnalysisHandler(workerPool, logger))      // analyze given position, up to 1 second for request
	router.HandleFunc("/rt-analysis", server.WsAnalysisHandler(workerPool, logger)) // real-time analysis only with websocket connection
	router.HandleFunc("/limits", server.LimitsHandler())                            // get current engine limits for the frontend
	router.HandleFunc("/health", server.HealthHandler(workerPool))                  // more in-depth health of the server
	router.HandleFunc("/healthz", server.HealthzHandler())                          // either 204 or 503 response
	router.HandleFunc("/metrics", server.MetricsHandler())                          // memory usage, pool usage and other stats

	srv := &http.Server{
		Addr:         ":" + server.DefaultConfig.Server.Port,
		Handler:      router,
		ReadTimeout:  server.DefaultConfig.Server.ReadTimeout,
		WriteTimeout: server.DefaultConfig.Server.WriteTimeout,
	}

	// Graceful shutdown, closes all remaning jobs on Ctrl+C
	done := make(chan bool)
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		logger.Info("Shutting down server...")
		workerPool.Stop()
		ctx, cancel := context.WithTimeout(ctx, server.DefaultConfig.Server.ShutdownTimeout)

		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("Couldn't shutdown server gracefully...")
		}
		close(done)
	}()

	ipAddress := server.GetOutboundIP()
	hostname := "localhost"
	if ipAddress != nil {
		hostname = ipAddress.String()
	}

	logger.Info(fmt.Sprintf(
		"Listening on: http://%s:%s", hostname, server.DefaultConfig.Server.Port,
	))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(err.Error())
	}

	<-done
	logger.Info("Server stopped")
}
