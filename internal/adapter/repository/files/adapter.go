package adapter

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	filesRepositoryAdapterPort "github.com/flash-go/files-service/internal/port/adapter/repository/files"
)

type Config struct {
	StoreLocalRootPath string
}

func New(config *Config) filesRepositoryAdapterPort.Interface {
	return &adapter{
		storeLocalRootPath: config.StoreLocalRootPath,
	}
}

type adapter struct {
	storeLocalRootPath string
}

func (a *adapter) CreateFile(ctx context.Context, data *filesRepositoryAdapterPort.CreateFileData) error {
	// Build path
	var path strings.Builder
	path.WriteString(a.storeLocalRootPath)
	if data.Path != "" {
		path.WriteString("/")
		path.WriteString(data.Path)
	}

	// Check dir exist
	if _, err := os.Stat(path.String()); err != nil {
		if os.IsNotExist(err) {
			return filesRepositoryAdapterPort.ErrDirNotFound
		}
	}

	// Build filename
	path.WriteString("/")
	path.WriteString(data.File.Filename)
	filename := path.String()

	// Check file exist
	if _, err := os.Stat(filename); err == nil {
		return filesRepositoryAdapterPort.ErrFileExist
	}

	// Open file
	src, err := data.File.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Create file
	dst, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

func (a *adapter) GetFiles(ctx context.Context, data *filesRepositoryAdapterPort.GetFilesData) (*[]filesRepositoryAdapterPort.FileResult, error) {
	// Build path
	var path strings.Builder
	path.WriteString(a.storeLocalRootPath)
	if data.Path != "" {
		path.WriteString("/")
		path.WriteString(data.Path)
	}
	fullPath := path.String()

	// Check dir exist
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return nil, filesRepositoryAdapterPort.ErrDirNotFound
		}
	}

	// Read dir
	files, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	// Build response
	response := make([]filesRepositoryAdapterPort.FileResult, len(files))
	for i, file := range files {
		info, err := file.Info()
		if err != nil {
			return nil, err
		}

		fileInfo := filesRepositoryAdapterPort.FileResult{
			Name:  file.Name(),
			IsDir: file.IsDir(),
		}

		if !file.IsDir() {
			s := info.Size()
			fileInfo.Size = &s

			f, err := os.Open(filepath.Join(fullPath, file.Name()))
			if err == nil {
				defer f.Close()
				buf := make([]byte, 512)
				n, _ := f.Read(buf)
				mt := http.DetectContentType(buf[:n])
				fileInfo.MimeType = &mt
			}
		}

		response[i] = fileInfo
	}

	// Sorting
	sort.Slice(response, func(i, j int) bool {
		if response[i].IsDir != response[j].IsDir {
			return response[i].IsDir
		}
		return response[i].Name < response[j].Name
	})

	return &response, nil
}

func (a *adapter) DeleteFile(ctx context.Context, data *filesRepositoryAdapterPort.DeleteFileData) error {
	// Build full path
	var path strings.Builder
	path.WriteString(a.storeLocalRootPath)
	path.WriteString("/")
	path.WriteString(data.Path)

	// Get path info
	if _, err := os.Stat(path.String()); err != nil {
		// Check dir exist
		if os.IsNotExist(err) {
			return filesRepositoryAdapterPort.ErrFileNotFound
		}
	}

	return os.Remove(path.String())
}

func (a *adapter) RenameFile(ctx context.Context, data *filesRepositoryAdapterPort.RenameFileData) error {
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

	// Check old file exist
	if _, err := os.Stat(oldPath.String()); err != nil && os.IsNotExist(err) {
		return filesRepositoryAdapterPort.ErrFileOldNotFound
	}

	// Check new file exist
	if _, err := os.Stat(newPath.String()); err == nil {
		return filesRepositoryAdapterPort.ErrFileNewExist
	}

	return os.Rename(
		oldPath.String(),
		newPath.String(),
	)
}
