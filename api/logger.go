package api

import (
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Logger logs API requests
func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		log.WithFields(log.Fields{
			"Method":      r.Method,
			"Request URI": r.RequestURI,
			"Name":        name,
			"Time":        time.Since(start),
		}).Debug("API request")
	})
}
