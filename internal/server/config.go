package server

import (
	"fmt"
	"time"
	"uttt/internal/mcts"
	"uttt/internal/utils"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

type Config struct {
	Server ServerConfig
	Pool   PoolConfig
	Engine EngineConfig
	Rate   RateLimitConfig
}

// Server config
type ServerConfig struct {
	AllowedOrigins  string // Access-Control-Allow-Origin header value, by default "*"
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// Pool config
type PoolConfig struct {
	DefaultWorkers   int           // number of 'engines' to run asynchornously
	DefaultQueueSize int           // number of possible requests
	JobTimeout       time.Duration // Time after which the analysis will be timedout
}

// Engine config
type EngineConfig struct {
	DefaultLimits mcts.Limits `json:"-"` // Default limits struct for the engine
	MaxDepth      int         `json:"depth"`
	MaxMovetime   int         `json:"-"`      // In milliseconds (10000 ms by default)
	MaxSizeMb     int         `json:"mbsize"` // maximum size of the tree in mb
	MaxMultiPv    int         `json:"multipv"`
	Threads       int         `json:"threads"` // number of threads to use by default
}

type RateLimitConfig struct {
	RequestsPerSecond rate.Limit
	Burst             int
}

// Default configs
var Version = "1.0.0"
var DefaultConfig Config

func LoadConfig() {
	// Load .env file (if exists)
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Failed to load .env file, using the default settings.")
	}

	DefaultConfig = Config{
		Server: ServerConfig{
			AllowedOrigins:  utils.GetEnv("ALLOWED_ORIGINS", "*"),
			Port:            utils.GetEnv("PORT", "8080"),
			ReadTimeout:     utils.GetEnvDuration("READ_TIMEOUT", 30*time.Second),
			WriteTimeout:    utils.GetEnvDuration("WRITE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: utils.GetEnvDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		Pool: PoolConfig{
			DefaultWorkers:   utils.GetEnvInt("WORKERS", 4),
			DefaultQueueSize: utils.GetEnvInt("QUEUE_SIZE", 100),
			JobTimeout:       utils.GetEnvDuration("JOB_TIMEOUT", 10*time.Second),
		},
		Engine: EngineConfig{
			DefaultLimits: *mcts.DefaultLimits(),
			MaxDepth:      utils.GetEnvInt("MAX_DEPTH", 14),
			MaxMovetime:   utils.GetEnvInt("MAX_MOVETIME", 9000),
			MaxSizeMb:     utils.GetEnvInt("MAX_TREE_SIZE_MB", 16),
			Threads:       utils.GetEnvInt("N_SEARCH_THREADS", 4),
			MaxMultiPv:    utils.GetEnvInt("MAX_MULTI_PV", 3),
		},
		Rate: RateLimitConfig{
			RequestsPerSecond: rate.Limit(utils.GetEnvInt("RATE_LIMIT_RPS", 5)),
			Burst:             utils.GetEnvInt("RATE_LIMIT_BURST", 8),
		},
	}

	DefaultConfig.Engine.DefaultLimits.
		SetMovetime(1000).
		SetMbSize(DefaultConfig.Engine.MaxSizeMb).
		SetThreads(DefaultConfig.Engine.Threads)
}
