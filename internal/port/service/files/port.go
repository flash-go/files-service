package port

import (
	"context"
	"mime/multipart"
)

type Interface interface {
	CreateFile(ctx context.Context, data *CreateFileData) error
	GetFiles(ctx context.Context, data *GetFilesData) (*[]FileResult, error)
	DeleteFile(ctx context.Context, data *DeleteFileData) error
	RenameFile(ctx context.Context, data *RenameFileData) error
}

// Args

type CreateFileData struct {
	Path string
	File *multipart.FileHeader
}

type GetFilesData struct {
	Path string
}

type DeleteFileData struct {
	Path string
}

type RenameFileData struct {
	OldPath string
	NewPath string
}

// Results

type FileResult struct {
	Name     string
	IsDir    bool
	Size     *int64
	MimeType *string
}
