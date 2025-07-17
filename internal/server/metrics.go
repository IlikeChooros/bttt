package server

import (
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

var ServerMetrics = &Metrics{}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ServerMetrics.RequestsTotal.Add(1)
			ServerMetrics.RequestsActive.Add(1)
			defer ServerMetrics.RequestsActive.Add(-1)

			next.ServeHTTP(w, r)
		})
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
