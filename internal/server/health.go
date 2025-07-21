package server

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

// Health status endpoint, shows memory used, worker pool status
type HealthStats struct {
	Status     string          `json:"status"`
	TimeStanp  time.Time       `json:"time_stamp"`
	Uptime     string          `json:"uptime"`
	WorkerPool WorkerPoolStats `json:"worker_pool"`
	Memory     MemoryStats     `json:"memory"`
}

type WorkerPoolStats struct {
	ActiveWorkers int   `json:"active_workers"`
	QueueCapacity int   `json:"queue_capacity"`
	ActiveJobs    int64 `json:"acitve_jobs"`
	PendingJobs   int64 `json:"pending_jobs"`
	RefusedJobs   int64 `json:"refused_jobs"`
}

type MemoryStats struct {
	AllocMb      uint64 `json:"alloc_mb"`
	TotalAllocMb uint64 `json:"total_alloc_mb"`
	SysMb        uint64 `json:"sys_mb"`
	NumGC        uint64 `json:"num_gc"`
}

var startTime = time.Now()
var ServerStatus atomic.Uint32

const (
	ServerStatusOK = iota
	ServerStatusShuttingDown
)

func HealthzHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ServerStatus.Load() == ServerStatusShuttingDown {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func HealthHandler(wp *WorkerPool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m)

		health := HealthStats{
			Status:    "healthy",
			TimeStanp: time.Now(),
			Uptime:    time.Since(startTime).String(),
			WorkerPool: WorkerPoolStats{
				ActiveWorkers: wp.workers,
				QueueCapacity: cap(wp.jobQueue),
				ActiveJobs:    wp.ActiveJobs(),
				PendingJobs:   wp.PendingJobs(),
				RefusedJobs:   wp.RefusedJobs(),
			},
			Memory: MemoryStats{
				AllocMb:      bytesToMB(m.Alloc),
				TotalAllocMb: bytesToMB(m.TotalAlloc),
				SysMb:        bytesToMB(m.Sys),
				NumGC:        bytesToMB(uint64(m.NumGC)),
			},
		}

		// Worker pool is overwhelemed
		if float32(health.WorkerPool.ActiveJobs)/float32(health.WorkerPool.QueueCapacity) > 0.8 {
			health.Status = "degraded"
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(health)
	}
}

func bytesToMB(bytes uint64) uint64 {
	return bytes >> 20
}
