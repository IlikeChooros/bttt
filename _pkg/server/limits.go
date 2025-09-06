package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func LimitsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simply return current engine config
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(DefaultConfig.Engine); err != nil {
			fmt.Println(err)
		}
	}
}
