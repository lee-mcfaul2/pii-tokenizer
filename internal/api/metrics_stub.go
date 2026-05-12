package api

import "net/http"

// promhttpHandler is replaced in Task 20 with the real Prometheus handler.
func promhttpHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("# metrics disabled in stub; replaced in Task 20\n"))
	})
}
