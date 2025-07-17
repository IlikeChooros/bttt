package server

import (
	"encoding/json"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	uttt "uttt/internal/engine"
)

type Metrics struct {
	RequestsTotal         atomic.Uint64
	RequestsActive        atomic.Int64
	AnalysisTotal         atomic.Uint64
	AnalysisErrors        atomic.Uint64
	AnalysisDurationMutex sync.Mutex
	AnalysisAvgDuration   int64 // in ms
}

var ServerMetrics = Metrics{}

// Returns server metrics middleware, increments stats each time new request comes
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ServerMetrics.RequestsTotal.Add(1)
			ServerMetrics.RequestsActive.Add(1)
			defer ServerMetrics.RequestsActive.Add(-1)

			next.ServeHTTP(w, r)
		})
}

// handles /metrics endpoint
func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		defer ServerMetrics.AnalysisDurationMutex.Unlock()

		ServerMetrics.AnalysisDurationMutex.Lock()
		metrics := map[string]any{
			"requests_total":    ServerMetrics.RequestsTotal.Load(),
			"requests_active":   ServerMetrics.RequestsActive.Load(),
			"analysis_total":    ServerMetrics.AnalysisTotal.Load(),
			"analysis_errors":   ServerMetrics.AnalysisErrors.Load(),
			"analysis_avg_time": ServerMetrics.AnalysisAvgDuration,
		}

		_ = json.NewEncoder(w).Encode(metrics)
	}
}

func (wp *WorkerPool) processJobWithMetrics(engine *uttt.Engine, req AnalysisRequest) {
	start := time.Now()

	// Update average duration
	defer func() {
		duration := time.Since(start).Milliseconds()

		ServerMetrics.AnalysisDurationMutex.Lock()
		ServerMetrics.AnalysisAvgDuration +=
			(duration - ServerMetrics.AnalysisAvgDuration) /
				int64(ServerMetrics.AnalysisTotal.Load())
		ServerMetrics.AnalysisDurationMutex.Unlock()
	}()

	// Close this channel
	defer close(req.Response)

	ServerMetrics.AnalysisTotal.Add(1)
	resp := wp.handleSearch(req, engine)

	if resp.Error != "" {
		ServerMetrics.AnalysisErrors.Add(1)
	}

	req.Response <- resp
}
