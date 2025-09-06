package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	// mine
)

func AnalysisHandler(workerPool *WorkerPool, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := NewAnalysisRequest(r, false)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		logger.Info("New post analysis", "req=", req.String())

		// Try submitting the request
		if !workerPool.Submit(req) {
			http.Error(w, "Server is too busy", http.StatusServiceUnavailable)
			return
		}

		// Now wait for the analysis
		select {
		case resp := <-req.Response:
			logger.Info("Sending response", "data", resp)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		case <-time.After(DefaultConfig.Pool.JobTimeout):
			http.Error(w, "Analysis timeout", http.StatusRequestTimeout)
		}
	}
}
