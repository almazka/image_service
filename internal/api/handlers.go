package api

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"image_service/internal/service"
	"log"
	"net/http"
	"strings"
)

// Handlers содержит обработчики HTTP запросов
type Handlers struct {
	uploadService *service.UploadService
}

// NewHandlers создает новый экземпляр Handlers
func NewHandlers(uploadService *service.UploadService) *Handlers {
	return &Handlers{
		uploadService: uploadService,
	}
}

// UploadHandler обрабатывает загрузку файлов
func (h *Handlers) UploadHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := h.getClientIP(r)
	log.Printf("Starting file upload from %s, User-Agent: %s", clientIP, r.UserAgent())

	// Ограничиваем размер файла
	r.Body = http.MaxBytesReader(w, r.Body, h.uploadService.MaxFileSize)

	// Парсим multipart форму
	if err := r.ParseMultipartForm(h.uploadService.MaxFileSize); err != nil {
		log.Printf("File too large from %s: %v", clientIP, err)
		h.sendErrorResponse(w, "File too large", http.StatusBadRequest)
		return
	}

	// Получаем файл из формы
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error getting file from form from %s: %v", clientIP, err)
		h.sendErrorResponse(w, "Error getting file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Загружаем файл через сервис
	result, err := h.uploadService.UploadFile(file, header)
	if err != nil {
		log.Printf("Error uploading file from %s: %v", clientIP, err)
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Логируем успешную загрузку
	log.Printf("File uploaded successfully from %s: %s (original: %s, type: %s, size: %d)",
		clientIP, result.Filename, header.Filename, result.Type, result.Size)

	h.sendSuccessResponse(w, result)
}

// FileHandler отдает файлы клиенту
func (h *Handlers) FileHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем имя файла из параметров маршрута
	fileName := chi.URLParam(r, "filename")
	if fileName == "" {
		h.sendErrorResponse(w, "File not specified", http.StatusBadRequest)
		return
	}

	// Получаем путь к файлу через сервис
	filePath, err := h.uploadService.GetFile(fileName)
	if err != nil {
		clientIP := h.getClientIP(r)
		log.Printf("File not found for %s: %s, error: %v", clientIP, fileName, err)
		h.sendErrorResponse(w, "File not found", http.StatusNotFound)
		return
	}

	// Логируем успешную выдачу файла
	clientIP := h.getClientIP(r)
	log.Printf("File served to %s: %s", clientIP, fileName)

	// Отдаем файл
	http.ServeFile(w, r, filePath)
}

// HealthHandler для проверки здоровья сервиса
func (h *Handlers) HealthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Health check from %s", h.getClientIP(r))

	h.sendSuccessResponse(w, map[string]interface{}{
		"status":  "healthy",
		"service": "file-upload-service",
		"version": "1.0.0",
	})
}

// getClientIP возвращает IP адрес клиента
func (h *Handlers) getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	return strings.Split(r.RemoteAddr, ":")[0]
}

// sendSuccessResponse отправляет успешный JSON ответ
func (h *Handlers) sendSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status": "success",
		"data":   data,
	}

	jsonResponse, _ := json.Marshal(response)
	w.Write(jsonResponse)
}

// sendErrorResponse отправляет JSON ответ с ошибкой
func (h *Handlers) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"status":  "error",
		"message": message,
		"code":    statusCode,
	}

	jsonResponse, _ := json.Marshal(response)
	w.Write(jsonResponse)
}
