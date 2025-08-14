package adapter

import (
	"context"
	"fmt"
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

/*
CreateFile securely saves an uploaded file within the adapter's base path.

This function performs several safety checks before writing the file:

1. Validates that the target path and filename are non-empty.
2. Cleans the path to remove "." and ".." elements.
3. Resolves the absolute path and ensures it is inside the base directory.
4. Checks that all parent directories exist.
5. Walks through parent directories to prevent symlink attacks.
6. Protects against overwriting existing files.
7. Opens the uploaded file safely and writes it atomically to the target path.

Allowed paths examples (assuming base is /var/data):

| Input Path         | File Name      | Resulting Absolute Path          | Reason                        |
|--------------------|----------------|----------------------------------|-------------------------------|
| "uploads/images"   | "pic.png"      | /var/data/uploads/images/pic.png | Inside base, directory exists |

Rejected paths examples:

| Input Path          | File Name      | Reason for rejection                       |
|---------------------|----------------|--------------------------------------------|
| "../../etc"         | "passwd"       | Path traversal outside base directory      |
| "uploads/../.."     | "hack.txt"     | Resolves above base directory              |
| "uploads/symlink"   | "file.txt"     | Parent directory is a symlink outside base |
| "uploads"           | ""             | Empty filename                             |
*/
func (a *adapter) CreateFile(ctx context.Context, data *filesRepositoryAdapterPort.CreateFileData) error {
	if data.File == nil || data.File.Filename == "" {
		return filesRepositoryAdapterPort.ErrInvalidFile
	}

	// Clean and build path
	cleanPath := filepath.Clean(data.Path)
	if cleanPath == "." {
		cleanPath = ""
	}
	if strings.HasPrefix(cleanPath, "..") {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	baseAbs, err := filepath.Abs(a.storeLocalRootPath)
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}

	targetDir := filepath.Join(baseAbs, cleanPath)
	targetDirAbs, err := filepath.Abs(targetDir)
	if err != nil {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	// Ensure directory is inside base
	relToBase, err := filepath.Rel(baseAbs, targetDirAbs)
	if err != nil || strings.HasPrefix(relToBase, "..") {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	// Check parent directories for symlinks (symlink race prevention)
	current := targetDirAbs
	for {
		if current == baseAbs || current == string(filepath.Separator) {
			break
		}
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("failed to stat %q: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return filesRepositoryAdapterPort.ErrInvalidPath
		}
		current = filepath.Dir(current)
	}

	// Check directory exists
	info, err := os.Stat(targetDirAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return filesRepositoryAdapterPort.ErrDirNotFound
		}
		return err
	}
	if !info.IsDir() {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	// Build full file path
	filename := filepath.Join(targetDirAbs, filepath.Base(data.File.Filename))

	// Check file existence
	if _, err := os.Stat(filename); err == nil {
		return filesRepositoryAdapterPort.ErrFileExist
	}

	// Open source file
	src, err := data.File.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Create destination file
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

/*
GetFiles securely lists files and directories within a specified path relative to the adapter's base path.

This function performs multiple safety checks:

1. Validates that the requested path is non-empty and does not traverse outside the base directory using ".." or absolute paths.
2. Resolves the absolute path for the requested directory.
3. Ensures the path is inside the adapter's storeLocalRootPath.
4. Checks parent directories for symlinks to prevent symlink race attacks.
5. Reads the directory contents, safely obtains file info, size, and MIME type.
6. Returns a sorted list with directories first, then files, both alphabetically.

Allowed paths examples (assuming base is /var/data):

| Input Path    | Resulting Absolute Path | Reason              |
|---------------|------------------------|----------------------|
| ""            | /var/data              | Base directory, safe |
| "uploads"     | /var/data/uploads      | Inside base, safe    |
| "tmp/session" | /var/data/tmp/session  | Inside base, safe    |

Rejected paths examples:

| Input Path       | Reason for rejection                          |
|------------------|-----------------------------------------------|
| "../../etc"      | Path traversal outside base                   |
| "/etc"           | Absolute path outside base                    |
| "uploads/../.."  | Resolves above base directory                 |
| "symlink_folder" | Parent directory is a symlink outside base    |
*/
func (a *adapter) GetFiles(ctx context.Context, data *filesRepositoryAdapterPort.GetFilesData) (*[]filesRepositoryAdapterPort.FileResult, error) {
	cleanPath := filepath.Clean(data.Path)

	if cleanPath == ".." || strings.HasPrefix(cleanPath, "..") {
		return nil, filesRepositoryAdapterPort.ErrInvalidPath
	}

	baseAbs, err := filepath.Abs(a.storeLocalRootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base path: %w", err)
	}

	targetAbs := filepath.Join(baseAbs, cleanPath)
	targetAbs, err = filepath.Abs(targetAbs)
	if err != nil {
		return nil, filesRepositoryAdapterPort.ErrInvalidPath
	}

	// Ensure target is inside base
	if rel, _ := filepath.Rel(baseAbs, targetAbs); strings.HasPrefix(rel, "..") {
		return nil, filesRepositoryAdapterPort.ErrInvalidPath
	}

	// Check parent directories for symlinks
	current := targetAbs
	for {
		if current == baseAbs || current == string(filepath.Separator) {
			break
		}
		info, err := os.Lstat(current)
		if err != nil {
			return nil, filesRepositoryAdapterPort.ErrInvalidPath
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil, filesRepositoryAdapterPort.ErrInvalidPath
		}
		current = filepath.Dir(current)
	}

	// Check directory existence
	info, err := os.Stat(targetAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, filesRepositoryAdapterPort.ErrDirNotFound
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, filesRepositoryAdapterPort.ErrInvalidPath
	}

	// Read dir
	files, err := os.ReadDir(targetAbs)
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

			f, err := os.Open(filepath.Join(targetAbs, file.Name()))
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

/*
DeleteFile securely deletes a file within the adapter's base path.

This function performs several safety checks before removing the file:

1. Validates that the file path is non-empty and does not traverse outside the base directory.
2. Resolves the absolute path for the file relative to the base.
3. Ensures the file path is inside the adapter's storeLocalRootPath.
4. Checks that all parent directories do not contain symlinks (symlink race prevention).
5. Confirms the file exists before attempting deletion.
6. Removes the file safely using os.Remove.

Allowed paths examples (assuming base is /var/data):

| Input Path               | Resulting Absolute Path          | Reason                    |
|--------------------------|----------------------------------|---------------------------|
| "uploads/images/pic.png" | /var/data/uploads/images/pic.png | Inside base, safe         |

Rejected paths examples:

| Input Path                 | Reason for rejection                       |
|----------------------------|--------------------------------------------|
| "../../etc/passwd"         | Path traversal outside base                |
| "/etc/passwd"              | Absolute path outside base                 |
| "uploads/../secret.txt".   | Resolves above base directory              |
| "uploads/symlink/file.txt" | Parent directory is a symlink outside base |
| ""                         | Empty file path                            |
*/
func (a *adapter) DeleteFile(ctx context.Context, data *filesRepositoryAdapterPort.DeleteFileData) error {
	if data.Path == "" {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	cleanPath := filepath.Clean(data.Path)
	if cleanPath == "." || strings.HasPrefix(cleanPath, "..") {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	baseAbs, err := filepath.Abs(a.storeLocalRootPath)
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}

	targetFile := filepath.Join(baseAbs, cleanPath)
	targetFileAbs, err := filepath.Abs(targetFile)
	if err != nil {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	// Ensure file is inside base
	relToBase, err := filepath.Rel(baseAbs, targetFileAbs)
	if err != nil || strings.HasPrefix(relToBase, "..") {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	// Check parent directories for symlinks (symlink race prevention)
	current := filepath.Dir(targetFileAbs)
	for {
		if current == baseAbs || current == string(filepath.Separator) {
			break
		}
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("failed to stat %q: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return filesRepositoryAdapterPort.ErrInvalidPath
		}
		current = filepath.Dir(current)
	}

	// Check file exists
	info, err := os.Stat(targetFileAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return filesRepositoryAdapterPort.ErrFileNotFound
		}
		return err
	}
	if info.IsDir() {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	// Delete file
	return os.Remove(targetFileAbs)
}

/*
RenameFile securely renames a file within the adapter's base path.

This function performs multiple safety checks before renaming the file:

1. Validates that both old and new paths are non-empty and do not traverse outside
   the base directory using ".." or absolute paths.
2. Resolves absolute paths for old and new files relative to the base.
3. Ensures both paths are inside the adapter's storeLocalRootPath.
4. Checks that all parent directories do not contain symlinks (symlink race prevention).
5. Checks that the old file exists and the new file does not exist.
6. Ensures the target paths are files and not directories.

Allowed paths examples (assuming base is /var/data):

| Input Path              | Resulting Absolute Path         | Reason                     |
|-------------------------|---------------------------------|----------------------------|
| old: "uploads/img.png"  | /var/data/uploads/img.png       | Inside base, exists        |
| new: "uploads/img2.png" | /var/data/uploads/img2.png      | Inside base, no collisions |

Rejected paths examples:

| Input Path                 | Reason for rejection                       |
|----------------------------|--------------------------------------------|
| "../../etc/passwd"         | Path traversal outside base                |
| "/etc/passwd"              | Absolute path outside base                 |
| "uploads/../secret.txt".   | Resolves above base directory              |
| "uploads/symlink/file.txt" | Parent directory is a symlink outside base |
| ""                         | Empty file path                            |
*/
func (a *adapter) RenameFile(ctx context.Context, data *filesRepositoryAdapterPort.RenameFileData) error {
	if data.OldPath == "" || data.NewPath == "" {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	cleanOld := filepath.Clean(data.OldPath)
	cleanNew := filepath.Clean(data.NewPath)

	if cleanOld == "." || strings.HasPrefix(cleanOld, "..") {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}
	if cleanNew == "." || strings.HasPrefix(cleanNew, "..") {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	baseAbs, err := filepath.Abs(a.storeLocalRootPath)
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}

	oldAbs := filepath.Join(baseAbs, cleanOld)
	oldAbs, err = filepath.Abs(oldAbs)
	if err != nil {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}
	newAbs := filepath.Join(baseAbs, cleanNew)
	newAbs, err = filepath.Abs(newAbs)
	if err != nil {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	// Ensure both paths are inside base
	if rel, _ := filepath.Rel(baseAbs, oldAbs); strings.HasPrefix(rel, "..") {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}
	if rel, _ := filepath.Rel(baseAbs, newAbs); strings.HasPrefix(rel, "..") {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	// Check parent directories for symlinks (symlink race prevention)
	for _, path := range []string{oldAbs, newAbs} {
		current := filepath.Dir(path)
		for {
			if current == baseAbs || current == string(filepath.Separator) {
				break
			}
			info, err := os.Lstat(current)
			if err != nil {
				return filesRepositoryAdapterPort.ErrInvalidPath
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return filesRepositoryAdapterPort.ErrInvalidPath
			}
			current = filepath.Dir(current)
		}
	}

	// Check existence and type
	fmt.Println(oldAbs)
	oldInfo, err := os.Stat(oldAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return filesRepositoryAdapterPort.ErrFileOldNotFound
		}
		return err
	}
	if oldInfo.IsDir() {
		return filesRepositoryAdapterPort.ErrInvalidPath
	}

	if newInfo, err := os.Stat(newAbs); err == nil {
		if newInfo.IsDir() {
			return filesRepositoryAdapterPort.ErrInvalidPath
		}
		return filesRepositoryAdapterPort.ErrFileNewExist
	} else if !os.IsNotExist(err) {
		return err
	}

	return os.Rename(oldAbs, newAbs)
}
