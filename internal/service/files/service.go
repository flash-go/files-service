package service

import (
	"context"

	filesRepositoryAdapterPort "github.com/flash-go/files-service/internal/port/adapter/repository/files"
	filesServicePort "github.com/flash-go/files-service/internal/port/service/files"
)

type Config struct {
	FilesRepository filesRepositoryAdapterPort.Interface
}

func New(config *Config) filesServicePort.Interface {
	return &service{
		config.FilesRepository,
	}
}

type service struct {
	filesRepository filesRepositoryAdapterPort.Interface
}

func (s *service) CreateFile(ctx context.Context, data *filesServicePort.CreateFileData) error {
	d := filesRepositoryAdapterPort.CreateFileData(*data)
	return s.filesRepository.CreateFile(ctx, &d)
}

func (s *service) GetFiles(ctx context.Context, data *filesServicePort.GetFilesData) (*[]filesServicePort.FileResult, error) {
	d := filesRepositoryAdapterPort.GetFilesData(*data)
	if files, err := s.filesRepository.GetFiles(ctx, &d); err != nil {
		return nil, err
	} else {
		f := make([]filesServicePort.FileResult, len(*files))
		for i, file := range *files {
			f[i] = filesServicePort.FileResult(file)
		}
		return &f, nil
	}
}

func (s *service) DeleteFile(ctx context.Context, data *filesServicePort.DeleteFileData) error {
	d := filesRepositoryAdapterPort.DeleteFileData(*data)
	return s.filesRepository.DeleteFile(ctx, &d)
}

func (s *service) RenameFile(ctx context.Context, data *filesServicePort.RenameFileData) error {
	d := filesRepositoryAdapterPort.RenameFileData(*data)
	return s.filesRepository.RenameFile(ctx, &d)
}
