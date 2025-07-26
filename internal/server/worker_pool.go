package server

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	uttt "uttt/internal/engine"
	"uttt/internal/mcts"
)

type AnalysisResponse struct {
	Depth int      `json:"depth"`
	Pv    []string `json:"pv"`
	Nps   uint64   `json:"nps"`
	Eval  string   `json:"eval"`
	Error string   `json:"error,omitempty"`
}

type AnalysisRequest struct {
	Position string                `json:"position"`
	Movetime int                   `json:"movetime"`
	Depth    int                   `json:"depth"`
	Threads  int                   `json:"threads"`
	SizeMb   int                   `json:"sizemb"`
	Response chan AnalysisResponse `json:"-"`
}

func (r AnalysisRequest) String() string {
	builder := strings.Builder{}
	if err := json.NewEncoder(&builder).Encode(r); err != nil {
		return "error"
	}
	return builder.String()
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
	for i := range wp.workers {
		wp.wg.Add(1)
		go wp.worker(i, ctx)
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

func (wp *WorkerPool) worker(id int, ctx context.Context) {
	// Main worker thread
	defer wp.wg.Done()

	// Each worker has it's own engine instance
	engine := uttt.NewEngine()

	for {
		select {
		case req := <-wp.jobQueue:
			// fmt.Println("Processing by", id)
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

func getAnalysisLimits(req AnalysisRequest) *mcts.Limits {
	limits := mcts.DefaultLimits()
	*limits = DefaultConfig.Engine.DefaultLimits

	if req.Movetime > 0 {
		limits.SetMovetime(min(req.Movetime, DefaultConfig.Engine.MaxMovetime))
	}

	if req.Depth > 0 {
		limits.SetDepth(min(req.Depth, DefaultConfig.Engine.MaxDepth))
	}

	if req.Threads > 0 {
		limits.SetThreads(min(req.Threads, DefaultConfig.Engine.Threads))
	}

	if req.SizeMb > 0 {
		limits.SetMbSize(min(req.SizeMb, DefaultConfig.Engine.MaxSizeMb))
	}

	return limits
}

// Handle engine search
func (wp *WorkerPool) handleSearch(req AnalysisRequest, engine *uttt.Engine) AnalysisResponse {
	// Decrement the counter
	defer wp.activeJobs.Add(-1)

	// Update jobs counters
	wp.pendingJobs.Add(-1)
	wp.activeJobs.Add(1)

	// Read the parameters
	notation := uttt.StartingPosition
	if req.Position != "" {
		notation = req.Position
	}

	// Sets new position, also resets the engine state
	if err := engine.SetNotation(notation); err != nil {
		return AnalysisResponse{
			Error: "Invalid position notation",
		}
	}

	// Get the limits
	limits := getAnalysisLimits(req)

	// Use here a timeout
	searchFinished := make(chan bool)
	go searchTimeout(searchFinished, engine)

	// Search, and return the result
	// fmt.Println("limits", *limits)
	engine.SetLimits(limits)
	result := engine.Think()
	close(searchFinished)

	// fmt.Println("search-result", result)
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
