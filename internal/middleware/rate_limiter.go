package middleware

import (
	"net/http"

	"image_service/internal/rate_limiter"
)

func RateLimit(limiter *rate_limiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == "/upload" && r.Method == "POST":
				if !limiter.CanUploadDownload() {
					http.Error(w, "Too many upload requests", http.StatusTooManyRequests)
					return
				}
				defer limiter.ReleaseUploadDownload()

			case r.URL.Path == "/files/" && r.Method == "GET":
				if !limiter.CanListFiles() {
					http.Error(w, "Too many list requests", http.StatusTooManyRequests)
					return
				}
				defer limiter.ReleaseListFiles()

			case len(r.URL.Path) > 7 && r.URL.Path[:7] == "/files/" && r.Method == "GET":
				if !limiter.CanUploadDownload() {
					http.Error(w, "Too many download requests", http.StatusTooManyRequests)
					return
				}
				defer limiter.ReleaseUploadDownload()
			}

			next.ServeHTTP(w, r)
		})
	}
}
