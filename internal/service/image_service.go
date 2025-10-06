package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// UploadService сервис для работы с загрузкой файлов
type UploadService struct {
	UploadDir    string
	AllowedTypes []string
	MaxFileSize  int64
}

// UploadResult результат загрузки файла
type UploadResult struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Type     string `json:"type"`
	URL      string `json:"url"`
}

// NewUploadService создает новый экземпляр UploadService
func NewUploadService(uploadDir string, allowedTypes []string, maxFileSize int64) *UploadService {
	return &UploadService{
		UploadDir:    uploadDir,
		AllowedTypes: allowedTypes,
		MaxFileSize:  maxFileSize,
	}
}

// UploadFile загружает файл на сервер
func (s *UploadService) UploadFile(file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	// Проверяем тип файла
	contentType, err := s.validateFileType(file)
	if err != nil {
		return nil, err
	}

	// Генерируем уникальное имя файла
	fileExt := filepath.Ext(header.Filename)
	fileName := s.generateFileName() + fileExt
	filePath := filepath.Join(s.UploadDir, fileName)

	// Сохраняем файл на диск
	if err := s.saveFile(file, filePath); err != nil {
		return nil, err
	}

	// Возвращаем результат
	return &UploadResult{
		Filename: fileName,
		Size:     header.Size,
		Type:     contentType,
		URL:      fmt.Sprintf("/files/%s", fileName),
	}, nil
}

// GetFile возвращает путь к файлу
func (s *UploadService) GetFile(filename string) (string, error) {
	filePath := filepath.Join(s.UploadDir, filepath.Base(filename))

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file not found: %s", filename)
	}

	return filePath, nil
}

// validateFileType проверяет тип файла
func (s *UploadService) validateFileType(file multipart.File) (string, error) {
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	// Возвращаем указатель чтения в начало
	file.Seek(0, 0)

	contentType := http.DetectContentType(buffer)

	// Проверяем разрешен ли тип
	for _, allowedType := range s.AllowedTypes {
		if contentType == allowedType {
			return contentType, nil
		}
	}

	return "", fmt.Errorf("file type not allowed: %s", contentType)
}

// saveFile сохраняет файл на диск
func (s *UploadService) saveFile(file multipart.File, filePath string) error {
	dst, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		// Удаляем частично записанный файл в случае ошибки
		os.Remove(filePath)
		return fmt.Errorf("error saving file: %w", err)
	}

	return nil
}

// generateFileName генерирует уникальное имя файла
func (s *UploadService) generateFileName() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
