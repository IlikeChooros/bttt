package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	// mine
	uttt "uttt/internal/engine"
	"uttt/internal/mcts"
)

func sseWrite(w http.ResponseWriter, data any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "data: %s\n\n", string(bytes))
	return nil
}

// Use SSE for real-time analysis
func SseAnalysisHandler(workerPool *WorkerPool, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Parse the request, but use query parameters
		req, err := NewAnalysisRequest(r, true)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "text/event-stream; charset=utf-8")
		h.Set("Cache-Control", "no-cache, no-transform")
		h.Set("Connection", "keep-alive")
		h.Set("X-Accel-Buffering", "no")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		if err := req.Validate(); err != nil {
			_ = sseWrite(w, AnalysisResponse{Error: err.Error()})
			flusher.Flush()
			return
		}

		turn, _ := uttt.ReadTurn(req.Position)
		searchCh := make(chan AnalysisResponse, DefaultConfig.Engine.MaxDepth+1)
		req.Listener.
			OnDepth(func(lts mcts.ListenerTreeStats[uttt.PosType]) {
				result := uttt.ToSearchResult(lts, turn)
				searchCh <- AnalysisResponse{
					Lines: ToAnalysisLine(result.Lines, result.Turn),
					Depth: result.Depth,
					Cps:   result.Cps,
					Final: false,
				}
			}).
			OnStop(func(_ mcts.ListenerTreeStats[uttt.PosType]) {
				close(searchCh)
			})

		if !workerPool.Submit(req) {
			_ = sseWrite(w, AnalysisResponse{Error: "Server busy"})
			flusher.Flush()
			return
		}

		// Stream loop
	Loop:
		for {
			select {
			case <-r.Context().Done():
				logger.Info("Client disconnected")
				break Loop
			case <-workerPool.RootCtx().Done():
				logger.Info("Server shutting down")
				break Loop
			case sr, ok := <-searchCh:
				if !ok {
					break Loop
				}
				if err := sseWrite(w, sr); err != nil {
					break Loop
				}
				flusher.Flush()
			case <-time.After(DefaultConfig.Pool.JobTimeout):
				_ = sseWrite(w, AnalysisResponse{Error: "Timeout"})
				flusher.Flush()
				break Loop
			}
		}

		// Final response if available
		select {
		case final := <-req.Response:
			final.Final = true
			_ = sseWrite(w, final)
			flusher.Flush()
		default:
		}
	}
}

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
