package server

import (
	"net/http"
	"uttt/internal/utils"
)

var ALLOWED_ORIGINS string = "*"

func SetupCors() {
	ALLOWED_ORIGINS = utils.GetEnv("ALLOWED_ORIGINS", "*")
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", ALLOWED_ORIGINS)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
