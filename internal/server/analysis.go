package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
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

	for {
		var wsRequest AnalysisRequest
		// Read the request
		if err := conn.ReadJSON(&wsRequest); err != nil {
			logger.Error(err.Error())
			return
		}

		// Validate request
		if err := wsRequest.Validate(); err != nil {
			logger.Error(err.Error())
			if e := conn.WriteJSON(AnalysisResponse{Error: err.Error()}); e != nil {
				logger.Error(e.Error())
			}
			return
		}

		logger.Info("rtAnalysis: new request", "data", wsRequest)

		// Setup listeners & buffered channel
		// resultCounter := atomic.Int32{}
		turn, _ := uttt.ReadTurn(wsRequest.Position)
		searchResults := make(chan AnalysisResponse, DefaultConfig.Engine.MaxDepth+1)
		wsRequest.Response = make(chan AnalysisResponse)
		wsRequest.Kill = make(chan bool)
		wsRequest.Listener.
			OnDepth(func(lts mcts.ListenerTreeStats[uttt.PosType]) {
				result := uttt.ToSearchResult(lts, turn)

				response := AnalysisResponse{
					Lines: ToAnalysisLine(result.Lines, result.Turn),
					Depth: result.Depth,
					Nps:   result.Nps,
					Final: false,
				}
				searchResults <- response // Put it into channel queue, don't waste preciouse search time
			}).
			OnStop(func(lts mcts.ListenerTreeStats[uttt.PosType]) {
				close(searchResults)
			})

		// Submit to worker pool
		if !workerPool.Submit(&wsRequest) {
			if err := conn.WriteJSON(AnalysisResponse{
				Error: "Failed to submit analysis, try again later",
			}); err != nil {
				logger.Error("RtAnalysis: failed to notify user on failed analysis submit")
			}
			return
		}

		// Read the responses
		for resp := range searchResults {
			logger.Info("Sending response", "data", resp)
			err := conn.WriteJSON(resp)
			if err != nil {
				logger.Error("RtAnalysis: failed finished analysis notify")
				return
			}
		}

		// Wait for the response
		select {
		case resp := <-wsRequest.Response: // final response
			logger.Info("Sending last response", "data", resp)
			err := conn.WriteJSON(resp)
			if err != nil {
				logger.Error("RtAnalysis: failed finished analysis notify")
				return
			}
		case <-time.After(DefaultConfig.Pool.JobTimeout):
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

func AnalysisHandler(workerPool *WorkerPool, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AnalysisRequest

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		logger.Info("New post analysis", "req=", req)

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
			logger.Info("Sending response", "data", resp)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		case <-time.After(DefaultConfig.Pool.JobTimeout):
			http.Error(w, "Analysis timeout", http.StatusRequestTimeout)
		}
	}
}
