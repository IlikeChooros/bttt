package server

import (
	"log"
	"net/http"
	"os"
)

// Get new logger instance, prints to standard output with http: prefix,
// current date and time
func NewLogger() *log.Logger {
	return log.New(os.Stdout, "http:", log.LstdFlags)
}

func LoggingMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestId, ok := r.Context().Value(RequestIDKey).(string)

				if !ok {
					requestId = "unknown"
				}

				logger.Println(requestId, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()

			next.ServeHTTP(w, r)
		})
	}
}
