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
	cm   *server.ConnManager
)

func InitOnce() {
	once.Do(func() {
		uttt.Init()
		server.InitAuth()
		server.LoadConfig()
		cm = server.NewConnManager()
		ctx := context.Background()
		pool = server.NewWorkerPool(
			server.DefaultConfig.Pool.DefaultWorkers,
			server.DefaultConfig.Pool.DefaultQueueSize,
			ctx,
		)
		pool.Start()
	})
}

func GetPool() *server.WorkerPool {
	InitOnce()
	return pool
}

func GetConnManager() *server.ConnManager {
	InitOnce()
	return cm
}
