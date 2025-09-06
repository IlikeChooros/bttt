package server

import (
	"encoding/json"
	"fmt"
	"time"
	"uttt/_pkg/mcts"
	"uttt/_pkg/utils"

	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
)

type Config struct {
	Server ServerConfig    `json:"server"`
	Pool   PoolConfig      `json:"pool"`
	Engine EngineConfig    `json:"engine"`
	Rate   RateLimitConfig `json:"rate"`
}

func (c Config) String() string {
	json, err := json.Marshal(c)
	if err != nil {
		return "Failed to marshal config"
	}
	return string(json)
}

// Server config
type ServerConfig struct {
	AllowedOrigins  string        `json:"allowed_origins"` // Access-Control-Allow-Origin header value, by default "*"
	Port            string        `json:"port"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
}

// Pool config
type PoolConfig struct {
	DefaultWorkers   int           `json:"default_workers"`    // number of 'engines' to run asynchornously
	DefaultQueueSize int           `json:"default_queue_size"` // number of possible requests
	JobTimeout       time.Duration `json:"job_timeout"`        // Time after which the analysis will be timedout
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
	RequestsPerSecond rate.Limit `json:"requests_per_second"`
	Burst             int        `json:"burst"`
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
			JobTimeout:       utils.GetEnvDuration("JOB_TIMEOUT", 5*time.Second),
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
