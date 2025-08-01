package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	// mine
	uttt "uttt/internal/engine"
	"uttt/internal/mcts"

	// 3rd party
	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		// fix in production
		return true
	},
	// Add some common websocket headers by default
	HandshakeTimeout: 10 * time.Second,
}

func rtAnalysis(workerPool *WorkerPool, conn *websocket.Conn, logger *slog.Logger) {
	defer conn.Close()
	var mx sync.Mutex

	for {
		var wsRequest AnalysisRequest
		// Read the request
		if err := conn.ReadJSON(&wsRequest); err != nil {
			logger.Error(err.Error())
			return
		}

		logger.Info("rtAnalysis: new position", "pos", wsRequest.Position)

		// Setup listeners & buffered channel
		resultCounter := atomic.Int32{}
		searchResults := make(chan AnalysisResponse, DefaultConfig.Engine.MaxDepth)
		wsRequest.Response = make(chan AnalysisResponse)
		wsRequest.Kill = make(chan bool)
		wsRequest.Listener.
			OnDepth(func(lts mcts.ListenerTreeStats[uttt.PosType]) {
				result := uttt.ToSearchResult(lts)
				pv := make([]string, len(result.Pv))
				for i := range result.Pv {
					pv[i] = result.Pv[i].String()
				}
				response := AnalysisResponse{
					Depth: result.Depth,
					Pv:    pv,
					Nps:   result.Nps,
					Eval:  result.StringValue(),
					Final: false,
				}

				resultCounter.Add(1)
				if resultCounter.Load() < int32(DefaultConfig.Engine.MaxDepth) {
					searchResults <- response // Put it into channel queue, don't waste preciouse search time
				}
			}).
			OnStop(func(lts mcts.ListenerTreeStats[uttt.PosType]) {
				if resultCounter.Load() != int32(DefaultConfig.Engine.MaxDepth) {
					close(searchResults)
				} else {
					close(searchResults)
				}

			})

		// Submit to worker pool
		if !workerPool.Submit(&wsRequest) {
			mx.Lock()
			if err := conn.WriteJSON(AnalysisResponse{
				Error: "Failed to submit analysis, try again later",
			}); err != nil {
				logger.Error("RtAnalysis: failed to notify user on failed analysis submit")
			}
			mx.Unlock()
			return
		}

		// Read the responses
		for resp := range searchResults {
			mx.Lock()
			logger.Info("Sending response", "data", resp)
			err := conn.WriteJSON(resp)
			mx.Unlock()
			if err != nil {
				logger.Error("RtAnalysis: failed finished analysis notify")
				return
			}
		}

		// Wait for the response
		select {
		case resp := <-wsRequest.Response: // final response
			mx.Lock()
			logger.Info("Sending last response", "data", resp)
			err := conn.WriteJSON(resp)
			mx.Unlock()
			if err != nil {
				logger.Error("RtAnalysis: failed finished analysis notify")
				return
			}
		case <-time.After(DefaultConfig.Pool.JobTimeout):
			mx.Lock()
			defer mx.Unlock()
			if err := conn.WriteJSON(AnalysisResponse{Error: "Analysis timeout"}); err != nil {
				logger.Error("RtAnalysis: failed analysis timeout notify")
			}
			return
		}
	}
}

func WsAnalysisHandler(workerPool *WorkerPool, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("WebSocket connection attempt", "remote", r.RemoteAddr, "headers", r.Header)

		conn, err := wsUpgrader.Upgrade(w, r, nil)

		if err != nil {
			logger.Error("WebSocket upgrade failed", "error", err)
			http.Error(w, "Failed to establish websocket: "+err.Error(), http.StatusBadRequest)
			return
		}

		logger.Info("WebSocket connection established", "remote", r.RemoteAddr)

		// Handle real-time analysis in parallel
		go rtAnalysis(workerPool, conn, logger)
	}
}

func AnalysisHandler(workerPool *WorkerPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AnalysisRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		fmt.Println("req=", req)

		req.Response = make(chan AnalysisResponse, 1)
		req.Kill = make(chan bool)

		// Try submitting the request
		if !workerPool.Submit(&req) {
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
