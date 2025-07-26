package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func AnalysisHandler(workerPool *WorkerPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AnalysisRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		fmt.Println("req=", req)

		req.Response = make(chan AnalysisResponse, 1)

		// Try submitting the request
		if !workerPool.Submit(req) {
			http.Error(w, "Server is too busy", http.StatusServiceUnavailable)
			return
		}

		// Now wait for the analysis
		select {
		case resp := <-req.Response:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		case <-time.After(DefaultConfig.Pool.JobTimeout):
			http.Error(w, "Analysis timeout", http.StatusRequestTimeout)
		}
	}
}
