package port

import (
	"context"
)

type Interface interface {
	CreateDir(ctx context.Context, data *CreateDirData) error
	DeleteDir(ctx context.Context, data *DeleteDirData) error
	RenameDir(ctx context.Context, data *RenameDirData) error
}

// Args

type CreateDirData struct {
	Path string
}

type DeleteDirData struct {
	Path string
}

type RenameDirData struct {
	OldPath string
	NewPath string
}
