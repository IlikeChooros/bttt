package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	uttt "uttt/internal/engine"
	"uttt/internal/mcts"
)

type AnalysisLine struct {
	Eval    string   `json:"eval"`
	AbsEval string   `json:"abseval,omitempty"`
	Pv      []string `json:"pv"`
}

// Simply convert pv moves to pv strings, and convert eval to a string
func ToAnalysisLine(engineLines []uttt.EngineLine, turn uttt.TurnType) []AnalysisLine {
	lines := make([]AnalysisLine, len(engineLines))
	for i := range len(engineLines) {
		lines[i].Eval = engineLines[i].StringValue(turn, false)
		lines[i].AbsEval = engineLines[i].StringValue(turn, true)
		lines[i].Pv = make([]string, len(engineLines[i].Pv))
		for j := range len(engineLines[i].Pv) {
			lines[i].Pv[j] = engineLines[i].Pv[j].String()
		}
	}
	return lines
}

type AnalysisResponse struct {
	Lines []AnalysisLine `json:"lines"`
	Depth int            `json:"depth"`
	Cps   uint32         `json:"cps"`
	Final bool           `json:"final"`
	Error string         `json:"error,omitempty"`
}

type AnalysisRequest struct {
	Position string                           `json:"position"`
	Movetime int                              `json:"movetime"`
	Depth    int                              `json:"depth"`
	Threads  int                              `json:"threads"`
	SizeMb   int                              `json:"sizemb"`
	MultiPv  int                              `json:"multipv"`
	Ctx      context.Context                  `json:"-"`
	Response chan AnalysisResponse            `json:"-"`
	Listener mcts.StatsListener[uttt.PosType] `json:"-"`
}

func NewAnalysisRequest(r *http.Request, useQuery bool) (*AnalysisRequest, error) {
	var req AnalysisRequest

	atoi := func(s string, def int) int {
		if s == "" {
			return def
		}
		var v int
		_, err := fmt.Sscanf(s, "%d", &v)
		if err != nil {
			return def
		}
		return v
	}

	if useQuery {
		q := r.URL.Query()
		req.Position = strings.ReplaceAll(
			strings.ReplaceAll(q.Get("position"), "n", "/"),
			"_", " ",
		)
		req.Movetime = atoi(q.Get("movetime"), 0)
		req.Depth = atoi(q.Get("depth"), 0)
		req.Threads = atoi(q.Get("threads"), 0)
		req.SizeMb = atoi(q.Get("sizemb"), 0)
		req.MultiPv = atoi(q.Get("multipv"), 0)
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return nil, fmt.Errorf("failed to decode request body: %w", err)
		}
	}

	req.Ctx = r.Context()
	req.Response = make(chan AnalysisResponse)
	req.Listener = mcts.StatsListener[uttt.PosType]{}
	return &req, nil
}

func (r *AnalysisRequest) Validate() error {
	validator := func(value int, min int, max int, name string) error {
		if value < min || value > max {
			return fmt.Errorf("Invalid %s value: %d, expected between %d-%d", name, value, min, max)
		}
		return nil
	}

	// Check if that's a 'default' request
	if r.Position != "" && r.Movetime == 0 && r.Depth == 0 && r.Threads == 0 && r.SizeMb == 0 && r.MultiPv == 0 {
		// Set default values
		r.Movetime = DefaultConfig.Engine.MaxMovetime
		r.Depth = DefaultConfig.Engine.MaxDepth
		r.Threads = DefaultConfig.Engine.Threads
		r.SizeMb = DefaultConfig.Engine.MaxSizeMb
		r.MultiPv = 1
		return nil
	}

	if err := validator(r.Movetime, 0, DefaultConfig.Engine.MaxMovetime, "movetime"); err != nil {
		return err
	}
	if err := validator(r.Depth, 1, DefaultConfig.Engine.MaxDepth, "depth"); err != nil {
		return err
	}
	if err := validator(r.Threads, 1, DefaultConfig.Engine.Threads, "threads"); err != nil {
		return err
	}
	if err := validator(r.SizeMb, 1, DefaultConfig.Engine.MaxSizeMb, "sizemb"); err != nil {
		return err
	}
	if err := validator(r.MultiPv, 1, DefaultConfig.Engine.MaxMultiPv, "multipv"); err != nil {
		return err
	}
	return nil
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
	jobQueue    chan *AnalysisRequest
	wg          sync.WaitGroup
	quit        chan struct{}
	activeJobs  atomic.Int64
	pendingJobs atomic.Int64
	refusedJobs atomic.Int64
	ctx         context.Context
}

func NewWorkerPool(workers int, queueSize int, ctx context.Context) *WorkerPool {
	wp := &WorkerPool{
		workers:  workers,
		jobQueue: make(chan *AnalysisRequest, queueSize),
		wg:       sync.WaitGroup{},
		quit:     make(chan struct{}),
		ctx:      ctx,
	}
	return wp
}

func (wp *WorkerPool) RootCtx() context.Context {
	return wp.ctx
}

// Start the worker pool, submit new requests with `Submit` method
func (wp *WorkerPool) Start() {
	for i := range wp.workers {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Wait's for all tasks to finish
func (wp *WorkerPool) Stop() {
	close(wp.quit)
	wp.wg.Wait()
}

// Try to submit new request
func (wp *WorkerPool) Submit(request *AnalysisRequest) bool {
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

func (wp *WorkerPool) worker(id int) {
	// Main worker thread
	defer wp.wg.Done()

	// Each worker has it's own engine instance
	engine := uttt.NewEngine()

	for {
		select {
		case req := <-wp.jobQueue:
			// fmt.Println("Processing by", id)
			wp.processJobWithMetrics(engine, req)
		case <-wp.ctx.Done():
			return
		case <-wp.quit:
			return
		}
	}
}

// A simple timeout funciton, call engine.Stop() after MaxMovetime elapses, else does nothing
func searchTimeout(ctx context.Context, req *AnalysisRequest, engine *uttt.Engine) {
	select {
	// Background context canceled
	case <-ctx.Done():
		engine.Stop()
		// Request context canceled
	case <-req.Ctx.Done():
		engine.Stop()
	case <-time.After(time.Duration(DefaultConfig.Engine.MaxMovetime) * time.Millisecond):
		engine.Stop()
	}
}

func getAnalysisLimits(req *AnalysisRequest) *mcts.Limits {
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

	limits.SetMultiPv(req.MultiPv)

	return limits
}

// Handle engine search
func (wp *WorkerPool) handleSearch(req *AnalysisRequest, engine *uttt.Engine) error {
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
		req.Response <- AnalysisResponse{
			Error: "Invalid position notation",
		}
		return fmt.Errorf("Invalid position notation: %s", notation)
	}

	// Get the limits and attach new listener
	limits := getAnalysisLimits(req)
	engine.Mcts().ResetListener()
	*engine.Mcts().StatsListener() = req.Listener

	// Use here a timeout
	go searchTimeout(wp.ctx, req, engine)

	// Search, and return the result
	engine.SetLimits(limits)
	result := engine.Think()

	// Set the response object
	req.Response <- AnalysisResponse{
		Lines: ToAnalysisLine(result.Lines, engine.Position().Turn()),
		Depth: result.Depth,
		Cps:   result.Cps,
		Final: true,
	}

	return nil
}
