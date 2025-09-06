package vercel

import (
	"context"
	"sync"
	uttt "uttt/_pkg/engine"
	server "uttt/_pkg/server"
)

var (
	once sync.Once
	pool *server.WorkerPool
)

func InitOnce() {
	once.Do(func() {
		uttt.Init()
		server.LoadConfig()
		ctx := context.Background()
		pool = server.NewWorkerPool(
			server.DefaultConfig.Pool.DefaultWorkers,
			server.DefaultConfig.Pool.DefaultQueueSize,
		)
		pool.Start(ctx)
	})
}

func GetPool() *server.WorkerPool {
	InitOnce()
	return pool
}
