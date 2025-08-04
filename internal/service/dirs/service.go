package service

import (
	"context"

	dirsRepositoryAdapterPort "github.com/flash-go/files-service/internal/port/adapter/repository/dirs"
	dirsServicePort "github.com/flash-go/files-service/internal/port/service/dirs"
)

type Config struct {
	DirsRepository dirsRepositoryAdapterPort.Interface
}

func New(config *Config) dirsServicePort.Interface {
	return &service{
		config.DirsRepository,
	}
}

type service struct {
	dirsRepository dirsRepositoryAdapterPort.Interface
}

func (s *service) CreateDir(ctx context.Context, data *dirsServicePort.CreateDirData) error {
	d := dirsRepositoryAdapterPort.CreateDirData(*data)
	return s.dirsRepository.CreateDir(ctx, &d)
}

func (s *service) DeleteDir(ctx context.Context, data *dirsServicePort.DeleteDirData) error {
	d := dirsRepositoryAdapterPort.DeleteDirData(*data)
	return s.dirsRepository.DeleteDir(ctx, &d)
}

func (s *service) RenameDir(ctx context.Context, data *dirsServicePort.RenameDirData) error {
	d := dirsRepositoryAdapterPort.RenameDirData(*data)
	return s.dirsRepository.RenameDir(ctx, &d)
}
