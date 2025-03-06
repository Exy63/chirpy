package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	res := fmt.Sprintf(`
	<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>
	</html>
`, cfg.fileserverHits.Load())
	w.Write([]byte(res))
}

func (cfg *apiConfig) handlerResetUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.platform == "dev" {
		cfg.dbQueries.ResetUsers(r.Context())
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}
