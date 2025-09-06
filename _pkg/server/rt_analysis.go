package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	uttt "uttt/_pkg/engine"
	"uttt/_pkg/mcts"
)

type SseAnalysisRequest struct {
	AnalysisRequest
	ConnId string `json:"connId"`
}

func SseSend(w http.ResponseWriter, event string, v any) {
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache, no-transform")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")

	data, err := json.Marshal(v)
	if err != nil {
		data = []byte(`{"error":"internal error"}`)
	}

	if event != "" {
		_, _ = w.Write([]byte("event: " + event + "\n"))
	}

	_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// On GET, register this user and connection ID, then return the initial
// SSE response headers, and keep the connection open for possible events
// It will send 3 types of events:
// 1. "connected" - sent once when the connection is established, with the connection ID
// 2. "analysis" - sent whenever there is a new analysis result
// 3. "ping" - sent every 30 seconds to keep the connection alive
func StableSSEHandler(cm *ConnManager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// auth the user
		userId, err := Authenticate(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// create a fresh connection ID
		connId := RandID(16)
		client := cm.Subscribe(userId, connId)
		defer cm.Unsubscribe(userId, connId)

		// Send a connected event
		SseSend(w, "connected", map[string]string{"connId": connId})

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				logger.Info("Client disconnected", "userId", userId, "connId", connId)
				return
			case <-client.ctx.Done():
				logger.Info("Server closed client", "userId", userId, "connId", connId)
				return
			case event, ok := <-client.Events:
				if !ok {
					logger.Info("Client events channel closed", "userId", userId, "connId", connId)
					return
				}
				SseSend(w, "analysis", event)
			case <-ticker.C:
				// Send a ping to keep the connection alive
				SseSend(w, "ping", map[string]string{"ts": fmt.Sprintf("%d", time.Now().Unix())})
			}
		}
	}
}

// On POST, submit an analysis request tied to this user and connection ID
func AnalysisSSESubmit(workerPool *WorkerPool, cm *ConnManager, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// auth the user
		userId, err := Authenticate(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get the connection ID from the body
		var sseReq SseAnalysisRequest
		if err := json.NewDecoder(r.Body).Decode(&sseReq); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		c := cm.Get(userId, sseReq.ConnId)

		if c == nil {
			http.Error(w, "Invalid connection ID", http.StatusBadRequest)
			return
		}

		// Parse the request, but use query parameters
		if err := sseReq.AnalysisRequest.Validate(); err != nil {
			http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
			return
		}

		turn, _ := uttt.ReadTurn(sseReq.Position)
		sseReq.Listener.
			OnDepth(func(lts mcts.ListenerTreeStats[uttt.PosType]) {
				result := uttt.ToSearchResult(lts, turn)

				cm.Publish(userId, sseReq.ConnId, AnalysisEvent{
					AnalysisResponse: AnalysisResponse{
						Lines: ToAnalysisLine(result.Lines, result.Turn),
						Depth: result.Depth,
						Cps:   result.Cps,
						Final: false,
					},
				})
			}).
			OnStop(func(lts mcts.ListenerTreeStats[uttt.PosType]) {
				result := uttt.ToSearchResult(lts, turn)

				cm.Publish(userId, sseReq.ConnId, AnalysisEvent{
					AnalysisResponse: AnalysisResponse{
						Lines: ToAnalysisLine(result.Lines, result.Turn),
						Depth: result.Depth,
						Cps:   result.Cps,
						Final: true,
					},
				})
			})

		// Make sure the 'Response' channel is not used
		sseReq.PublishLastWithStop = true
		if !workerPool.Submit(&sseReq.AnalysisRequest) {
			http.Error(w, "Server is too busy", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
