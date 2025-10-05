package main

import (
	"fmt"
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	resp := fmt.Sprintf(`<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
</html>`, cfg.fileserverHits.Load())
	w.Write([]byte(resp))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	cfg.handlerHits(w, r)
	err := cfg.db.Reset(r.Context())
	if err != nil {
		log.Printf("Error resetting the users table: %s", err)
	}
}
