package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handlerHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	resp := fmt.Sprintf("Hits: %v", cfg.fileserverHits.Load())
	w.Write([]byte(resp))
}

func (cfg *apiConfig) handlerHitsReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	cfg.handlerHits(w, r)
}
