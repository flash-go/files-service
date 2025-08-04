package adapter

import (
	"context"
	"os"
	"strings"

	dirsRepositoryAdapterPort "github.com/flash-go/files-service/internal/port/adapter/repository/dirs"
)

type Config struct {
	StoreLocalRootPath string
}

func New(config *Config) dirsRepositoryAdapterPort.Interface {
	return &adapter{
		storeLocalRootPath: config.StoreLocalRootPath,
	}
}

type adapter struct {
	storeLocalRootPath string
}

func (a *adapter) CreateDir(ctx context.Context, data *dirsRepositoryAdapterPort.CreateDirData) error {
	// Build full path
	var path strings.Builder
	path.WriteString(a.storeLocalRootPath)
	path.WriteString("/")
	path.WriteString(data.Path)

	// Get path info
	if _, err := os.Stat(path.String()); err != nil {
		// Check dir exist
		if os.IsExist(err) {
			return dirsRepositoryAdapterPort.ErrDirExist
		}
	}

	return os.MkdirAll(path.String(), 0700)
}

func (a *adapter) DeleteDir(ctx context.Context, data *dirsRepositoryAdapterPort.DeleteDirData) error {
	// Build full path
	var path strings.Builder
	path.WriteString(a.storeLocalRootPath)
	path.WriteString("/")
	path.WriteString(data.Path)

	// Get path info
	if _, err := os.Stat(path.String()); err != nil {
		// Check dir exist
		if os.IsNotExist(err) {
			return dirsRepositoryAdapterPort.ErrDirNotFound
		}
	}

	return os.RemoveAll(path.String())
}

func (a *adapter) RenameDir(ctx context.Context, data *dirsRepositoryAdapterPort.RenameDirData) error {
	// Build old full path
	var oldPath strings.Builder
	oldPath.WriteString(a.storeLocalRootPath)
	oldPath.WriteString("/")
	oldPath.WriteString(data.OldPath)

	// Build new full path
	var newPath strings.Builder
	newPath.WriteString(a.storeLocalRootPath)
	newPath.WriteString("/")
	newPath.WriteString(data.NewPath)

	// Check dir exist
	if _, err := os.Stat(oldPath.String()); err != nil && os.IsNotExist(err) {
		return dirsRepositoryAdapterPort.ErrDirOldNotFound
	}

	// Check dir exist
	if _, err := os.Stat(newPath.String()); err == nil {
		return dirsRepositoryAdapterPort.ErrDirNewExist
	}

	return os.Rename(
		oldPath.String(),
		newPath.String(),
	)
}
