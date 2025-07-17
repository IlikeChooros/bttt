package server

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

const (
	RequestIDTag = "X-Request-Id"
)

type RequestIDType struct {
	id int
}

var RequestIDKey = RequestIDType{5}

func nextRequestId() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func TracingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestId := r.Header.Get(RequestIDTag)
			if requestId == "" {
				requestId = nextRequestId()
			}

			// Attach request ID to context and the incoming request
			context := context.WithValue(r.Context(), RequestIDKey, requestId)
			w.Header().Set(RequestIDTag, requestId)
			next.ServeHTTP(w, r.WithContext(context))
		})
	}
}
