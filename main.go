package main

import (
	"log"
	"net/http"
	"os"

	"image_service/internal/api"
	"image_service/internal/service"

	"gopkg.in/yaml.v3"
)

// Config структура для конфигурации
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
}

type ServerConfig struct {
	Port        string `yaml:"port"`
	MaxFileSize int64  `yaml:"max_file_size"`
}

type StorageConfig struct {
	UploadDir    string   `yaml:"upload_dir"`
	AllowedTypes []string `yaml:"allowed_types"`
}

func main() {
	// Загружаем конфигурацию
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	// Создаем директорию для загрузок
	if err := os.MkdirAll(config.Storage.UploadDir, 0755); err != nil {
		log.Fatalf("Error creating upload directory %s: %v", config.Storage.UploadDir, err)
	}

	// Инициализируем сервисы
	uploadService := service.NewUploadService(
		config.Storage.UploadDir,
		config.Storage.AllowedTypes,
		config.Server.MaxFileSize,
	)

	// Создаем роутер
	router := api.NewRouter(uploadService)

	// Запускаем сервер
	log.Printf("Server starting on port %s", config.Server.Port)
	log.Printf("Upload directory: %s", config.Storage.UploadDir)
	log.Printf("Allowed file types: %v", config.Storage.AllowedTypes)
	log.Printf("Max file size: %d bytes", config.Server.MaxFileSize)

	if err := http.ListenAndServe(config.Server.Port, router); err != nil {
		log.Fatal("Error starting server:", err)
	}
}

// loadConfig загружает конфигурацию из YAML файла
func loadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
