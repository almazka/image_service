package main

import (
	"context"
	"image_service/internal/middleware"
	"image_service/internal/rate_limiter"
	"log"
	"net"
	"net/http"
	"os"

	"image_service/internal/api"
	grpcserver "image_service/internal/grpc"
	"image_service/internal/service"
	"image_service/proto"

	"google.golang.org/grpc"
	"gopkg.in/yaml.v3"
)

// Config структура для конфигурации
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
	GRPC    GRPCConfig    `yaml:"grpc"`
}

type ServerConfig struct {
	Port        string `yaml:"port"`
	MaxFileSize int64  `yaml:"max_file_size"`
}

type StorageConfig struct {
	UploadDir    string   `yaml:"upload_dir"`
	AllowedTypes []string `yaml:"allowed_types"`
}

type GRPCConfig struct {
	Port string `yaml:"port"`
}

var (
	rateLimiter *rate_limiter.RateLimiter
)

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

	// Создаем rate limiter (10 upload/download, 100 list files)
	rateLimiter = rate_limiter.New(10, 100)

	// Инициализируем сервисы
	uploadService := service.NewUploadService(
		config.Storage.UploadDir,
		config.Storage.AllowedTypes,
		config.Server.MaxFileSize,
	)

	log.Printf("Starting image_service...")
	log.Printf("Upload directory: %s", config.Storage.UploadDir)
	log.Printf("Allowed file types: %v", config.Storage.AllowedTypes)
	log.Printf("Max file size: %d bytes", config.Server.MaxFileSize)
	log.Printf("HTTP server port: %s", config.Server.Port)
	log.Printf("gRPC server port: %s", config.GRPC.Port)

	// Запускаем HTTP сервер в горутине
	go startHTTPServer(uploadService, config.Server.Port)

	// Запускаем gRPC сервер в основной горутине
	startGRPCServer(uploadService, config.GRPC.Port)
}

// startHTTPServer запускает HTTP сервер
func startHTTPServer(uploadService *service.UploadService, port string) {
	// Создаем роутер
	router := api.NewRouter(uploadService)

	handler := middleware.RateLimit(rateLimiter)(router)

	log.Printf("HTTP Server starting on port %s", port)

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}

// startGRPCServer запускает gRPC сервер
func startGRPCServer(uploadService *service.UploadService, port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			switch info.FullMethod {
			case "/image_service.FileService/UploadFile":
				if !rateLimiter.CanUploadDownload() {
					return nil, grpc.Errorf(grpc.Code(nil), "too many upload requests")
				}
				defer rateLimiter.ReleaseUploadDownload()

			case "/image_service.FileService/ListFiles":
				if !rateLimiter.CanListFiles() {
					return nil, grpc.Errorf(grpc.Code(nil), "too many list requests")
				}
				defer rateLimiter.ReleaseListFiles()
			}

			return handler(ctx, req)
		}),

		grpc.StreamInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			if info.FullMethod == "/image_service.FileService/DownloadFile" {
				if !rateLimiter.CanUploadDownload() {
					return grpc.Errorf(grpc.Code(nil), "too many download requests")
				}
				defer rateLimiter.ReleaseUploadDownload()
			}

			return handler(srv, ss)
		}),
	)

	fileService := grpcserver.NewServer(uploadService)
	proto.RegisterFileServiceServer(grpcServer, fileService)

	log.Printf("gRPC server starting on %s", port)
	grpcServer.Serve(lis)
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
