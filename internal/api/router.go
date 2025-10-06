package api

import (
	"net/http"

	"image_service/internal/service"

	"github.com/go-chi/chi/v5"
)

// NewRouter создает новый роутер с настройками
func NewRouter(uploadService *service.UploadService) http.Handler {
	handlers := NewHandlers(uploadService)

	r := chi.NewRouter()

	// Routes
	r.Get("/health", handlers.HealthHandler)
	r.Post("/upload", handlers.UploadHandler)

	// File routes
	r.Route("/files", func(r chi.Router) {
		r.Get("/{filename}", handlers.FileHandler)
	})

	return r
}
