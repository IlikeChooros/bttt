package server

import (
	"fmt"
	"time"
	uttt "uttt/internal/engine"
	"uttt/internal/utils"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
	Pool   PoolConfig
	Engine EngineConfig
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
	DefaultLimits *uttt.Limits // Default limits struct for the engine
	MaxDepth      int
	MaxMovetime   int // In milliseconds (10000 ms by default)
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
			JobTimeout:       utils.GetEnvDuration("JOB_TIMEOUT", 30*time.Second),
		},
		Engine: EngineConfig{
			DefaultLimits: uttt.DefaultLimits().SetMovetime(1000),
			MaxDepth:      utils.GetEnvInt("MAX_DEPTH", 20),
			MaxMovetime:   utils.GetEnvInt("MAX_MOVETIME", 20000),
		},
	}
}
