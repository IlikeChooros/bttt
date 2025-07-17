package server

import (
	"context"
	"sync"
	uttt "uttt/internal/engine"
)

type AnalysisResponse struct {
	Pv    []string `json:"pv"`
	Nps   uint64   `json:"nps"`
	Eval  string   `json:"eval"`
	Error string   `json:"error,omitempty"`
}

type AnalysisRequest struct {
	Position string `json:"position"`
	Movetime int    `json:"movetime"`
	Depth    int    `json:"depth"`
	Response chan AnalysisResponse
}

type WorkerPool struct {
	workers  int
	jobQueue chan AnalysisRequest
	wg       sync.WaitGroup
	quit     chan struct{}
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	wp := &WorkerPool{
		workers:  workers,
		jobQueue: make(chan AnalysisRequest, queueSize),
		wg:       sync.WaitGroup{},
		quit:     make(chan struct{}),
		ctx:      context.Background(),
	}
	wp.ctx, wp.cancel = context.WithCancel(wp.ctx)
	return wp
}

func (wp *WorkerPool) Context() context.Context {
	return wp.ctx
}

// Start the worker pool, submit new requests with `Submit` method
func (wp *WorkerPool) Start() {
	for range wp.workers {
		wp.wg.Add(1)
		go wp.worker()
	}
}

// Wait's for all tasks to finish
func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.quit)
	wp.wg.Wait()
}

// Try to submit new request
func (wp *WorkerPool) Submit(request AnalysisRequest) bool {
	select {
	case wp.jobQueue <- request:
		return true
	default:
		return false
	}
}

func (wp *WorkerPool) worker() {
	// Main worker thread
	defer wp.wg.Done()

	// Each worker has it's own engine instance
	engine := uttt.NewEngine()

	for {
		select {
		case req := <-wp.jobQueue:
			wp.processJobWithMetrics(engine, req)
		case <-wp.ctx.Done():
			return
		case <-wp.quit:
			return
		}
	}
}

// Handle engine search
func (wp *WorkerPool) handleSearch(req AnalysisRequest, engine *uttt.Engine) AnalysisResponse {
	// Clear the search cache
	engine.NewGame()

	// Read the parameters
	notation := uttt.StartingPosition
	if req.Position != "" {
		notation = req.Position
	}

	// Invalid position
	if err := engine.Position().FromNotation(notation); err != nil {
		return AnalysisResponse{
			Error: "Invalid position notation",
		}
	}

	// Set the limits
	limits := uttt.DefaultLimits()
	// No limits provided
	if req.Depth == 0 && req.Movetime == 0 {
		limits.SetMovetime(1000)
	}

	if req.Movetime > 0 {
		limits.SetMovetime(min(req.Movetime, 30*1000))
	}

	if req.Depth > 0 {
		limits.SetDepth(min(req.Depth, uttt.MaxDepth-1))
	}

	// Search, and return the result
	engine.SetLimits(*limits)
	result := engine.Think(false)

	// Create the pv string slice
	pv := make([]string, engine.Pv().Size())
	slice := engine.Pv().Slice()
	for i := range slice {
		pv[i] = slice[i].String()
	}

	// Set the response object
	return AnalysisResponse{
		Pv:   pv,
		Nps:  result.Nps,
		Eval: result.String(),
	}
}
