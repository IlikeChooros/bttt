package utils

import (
	"os"
	"strconv"
	"time"
)

// Generic function to parse any env variable
// If no error occurs and value is not empty, return the parser's value
// Else returns defaultValue
func GetEnvT[T any](key string, defaultValue T, parser func(string) (T, error)) T {
	if value := os.Getenv(key); value != "" {
		if val, err := parser(key); err == nil {
			return val
		}
	}
	return defaultValue
}

func GetEnv(key, defaultValue string) string {
	return GetEnvT(key, defaultValue, func(s string) (string, error) { return defaultValue, nil })
}

func GetEnvInt(key string, defaultValue int) int {
	return GetEnvT(key, defaultValue, strconv.Atoi)
}

func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	return GetEnvT(key, defaultValue, time.ParseDuration)
}
