package server

import (
	"log/slog"
	"net/http"
	"os"
)

func NewLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				// requestId, ok := r.Context().Value(RequestIDKey).(string)
				// if !ok {
				// 	requestId = "unknown"
				// }

				// Standard log
				logger.Info(
					r.UserAgent(), r.Method, r.URL.Path,
					r.RemoteAddr, GetRequestIPAddress(r),
				)
			}()

			next.ServeHTTP(w, r)
		})
	}
}
