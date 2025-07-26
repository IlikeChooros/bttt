package server

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	uttt "uttt/internal/engine"
)

type AnalysisResponse struct {
	Depth int      `json:"depth"`
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
	workers     int
	jobQueue    chan AnalysisRequest
	wg          sync.WaitGroup
	quit        chan struct{}
	activeJobs  atomic.Int64
	pendingJobs atomic.Int64
	refusedJobs atomic.Int64
}

func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	wp := &WorkerPool{
		workers:  workers,
		jobQueue: make(chan AnalysisRequest, queueSize),
		wg:       sync.WaitGroup{},
		quit:     make(chan struct{}),
	}
	return wp
}

// Start the worker pool, submit new requests with `Submit` method
func (wp *WorkerPool) Start(ctx context.Context) {
	for range wp.workers {
		wp.wg.Add(1)
		go wp.worker(ctx)
	}
}

// Wait's for all tasks to finish
func (wp *WorkerPool) Stop() {
	close(wp.quit)
	wp.wg.Wait()
}

// Try to submit new request
func (wp *WorkerPool) Submit(request AnalysisRequest) bool {
	select {
	case wp.jobQueue <- request:
		wp.pendingJobs.Add(1)
		return true
	default:
		wp.refusedJobs.Add(1)
		return false
	}
}

// Get the number of waiting jobs on the queue
func (wp *WorkerPool) PendingJobs() int64 {
	return wp.pendingJobs.Load()
}

// Get the number of running analyses
func (wp *WorkerPool) ActiveJobs() int64 {
	return wp.pendingJobs.Load()
}

// Get the number of failed submits
func (wp *WorkerPool) RefusedJobs() int64 {
	return wp.refusedJobs.Load()
}

func (wp *WorkerPool) worker(ctx context.Context) {
	// Main worker thread
	defer wp.wg.Done()

	// Each worker has it's own engine instance
	engine := uttt.NewEngine()

	for {
		select {
		case req := <-wp.jobQueue:
			wp.processJobWithMetrics(engine, req)
		case <-ctx.Done():
			return
		case <-wp.quit:
			return
		}
	}
}

// A simple timeout funciton, call engine.Stop() after MaxMovetime elapses, else does nothing
func searchTimeout(done chan bool, engine *uttt.Engine) {
	select {
	case <-done:
		return
	case <-time.After(time.Duration(DefaultConfig.Engine.MaxMovetime) * time.Millisecond):
		engine.Stop()
	}
}

// Handle engine search
func (wp *WorkerPool) handleSearch(req AnalysisRequest, engine *uttt.Engine) AnalysisResponse {
	// Decrement the counter
	defer wp.activeJobs.Add(-1)

	// Update jobs counters
	wp.pendingJobs.Add(-1)
	wp.activeJobs.Add(1)

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
	limits := DefaultConfig.Engine.DefaultLimits
	if req.Movetime > 0 {
		limits.SetMovetime(min(req.Movetime, DefaultConfig.Engine.MaxMovetime))
	}

	if req.Depth > 0 {
		limits.SetDepth(min(req.Depth, DefaultConfig.Engine.MaxDepth))
	}

	// Use here a timeout
	searchFinished := make(chan bool)
	go searchTimeout(searchFinished, engine)

	// Search, and return the result
	engine.SetLimits(limits)
	result := engine.Think(false)
	close(searchFinished)

	// Create the pv string slice
	pv := make([]string, engine.Pv().Size())
	slice := engine.Pv().Slice()
	for i := range slice {
		pv[i] = slice[i].String()
	}

	// Set the response object
	return AnalysisResponse{
		Depth: result.Depth,
		Pv:    pv,
		Nps:   result.Nps,
		Eval:  result.StringValue(),
	}
}
