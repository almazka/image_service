package grpc

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"time"

	"image_service/internal/service"
	"image_service/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	proto.UnimplementedFileServiceServer
	uploadService *service.UploadService
}

func NewServer(uploadService *service.UploadService) *Server {
	return &Server{
		uploadService: uploadService,
	}
}

func (s *Server) UploadFile(ctx context.Context, req *proto.UploadFileRequest) (*proto.UploadFileResponse, error) {
	log.Printf("gRPC: Upload file %s (%d bytes)", req.Filename, len(req.Content))

	// Сохраняем файл
	filePath := filepath.Join(s.uploadService.UploadDir, req.Filename)
	err := os.WriteFile(filePath, req.Content, 0644)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to save file: %v", err)
	}

	return &proto.UploadFileResponse{
		Filename: req.Filename,
		Size:     int64(len(req.Content)),
		Url:      "/files/" + req.Filename,
		Message:  "File uploaded successfully via gRPC",
	}, nil
}

func (s *Server) ListFiles(ctx context.Context, req *proto.ListFilesRequest) (*proto.ListFilesResponse, error) {
	log.Printf("gRPC: List files request")

	files, err := s.uploadService.ListFiles()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list files: %v", err)
	}

	protoFiles := make([]*proto.FileInfo, len(files))
	for i, file := range files {
		protoFiles[i] = &proto.FileInfo{
			Name:      file.Name,
			Size:      file.Size,
			Url:       file.URL,
			CreatedAt: file.CreatedAt.Format(time.RFC3339),
			UpdatedAt: file.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &proto.ListFilesResponse{
		Files: protoFiles,
		Total: int32(len(protoFiles)),
	}, nil
}

func (s *Server) DownloadFile(req *proto.DownloadFileRequest, stream proto.FileService_DownloadFileServer) error {
	log.Printf("gRPC: Download file %s", req.Filename)

	filePath, err := s.uploadService.GetFile(req.Filename)
	if err != nil {
		return status.Errorf(codes.NotFound, "file not found: %s", req.Filename)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to read file: %v", err)
	}

	// Отправляем файл чанками по 64KB
	chunkSize := 64 * 1024
	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}

		err := stream.Send(&proto.DownloadFileResponse{
			Chunk: content[i:end],
		})
		if err != nil {
			return status.Errorf(codes.Internal, "failed to send chunk: %v", err)
		}
	}

	return nil
}

func (s *Server) HealthCheck(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	log.Printf("gRPC: Health check")
	return &proto.HealthCheckResponse{
		Status:  "healthy",
		Service: "image_service",
		Version: "1.0.0",
	}, nil
}
