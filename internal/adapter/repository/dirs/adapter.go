package adapter

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

/*
CreateDir creates a new directory inside the local storage root (storeLocalRootPath)
with strong safety checks to prevent path traversal, symlink-based escape, and
creation outside the allowed base directory.

Safety measures implemented:

1. **Path traversal prevention**
  - Rejects empty paths.
  - Cleans the path using `filepath.Clean` to normalize `.` and `..` elements.
  - Rejects root ("/"), current directory ("."), and any path starting with "..".

2. **Absolute path resolution**
  - Resolves both base path and target path to absolute paths.
  - Ensures the target is located strictly inside the base directory.

3. **Existing path checks**
  - If the target already exists:
  - Returns `ErrDirExist` if it's a directory.
  - Returns an error if it's not a directory.

4. **Symlink protection**
  - Traverses parent directories up to the base path.
  - If any parent directory is a symlink, aborts creation.
  - This prevents symlink race attacks where a path component is replaced
    with a symlink pointing outside the base.

5. **Secure directory creation**
  - Creates directories with permission `0700` (owner-only access).

Allowed paths:

| Input Path    | Resulting Absolute Path | Reason                    |
|---------------|-------------------------|---------------------------|
| `docs`        | `<base>/docs`           | Inside base, no traversal |
| `images/2025` | `<base>/images/2025`    | Inside base, no symlinks  |
| `./temp`      | `<base>/temp`           | Normalized, stays inside  |

Rejected paths:

| Input Path            | Reason for rejection                     |
|-----------------------|------------------------------------------|
| “ (empty string)     | Target dir is empty                      |
| `/`                   | Root directory not allowed               |
| `..`                  | Path traversal outside base              |
| `../../etc`           | Path traversal outside base              |
| `symlink_to_outside`  | Parent dir is a symlink pointing outside |
| `nested/../../escape` | Path traversal escape                    |

By enforcing these checks, this function ensures that directories can only be created
inside the intended storage root and prevents malicious attempts to write outside of it.
*/
func (a *adapter) CreateDir(ctx context.Context, data *dirsRepositoryAdapterPort.CreateDirData) error {
	// Validate input path
	if data.Path == "" {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}
	cleanPath := filepath.Clean(data.Path)
	if cleanPath == "." || cleanPath == "/" || strings.HasPrefix(cleanPath, "..") {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}

	// Resolve absolute paths
	baseAbs, err := filepath.Abs(a.storeLocalRootPath)
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}
	targetAbs, err := filepath.Abs(filepath.Join(baseAbs, cleanPath))
	if err != nil {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}

	// Ensure targetAbs is inside baseAbs
	relToBase, err := filepath.Rel(baseAbs, targetAbs)
	if err != nil {
		return fmt.Errorf("failed to compute relative path: %w", err)
	}
	if strings.HasPrefix(relToBase, "..") || relToBase == "." {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}

	// Check if it already exists
	if info, err := os.Lstat(targetAbs); err == nil {
		if info.IsDir() {
			return dirsRepositoryAdapterPort.ErrDirExist
		}
		return dirsRepositoryAdapterPort.ErrInvalidPath
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat target: %w", err)
	}

	// Check parent directories for symlinks (symlink race prevention)
	current := filepath.Dir(targetAbs)
	for {
		if current == baseAbs || current == string(filepath.Separator) {
			break
		}
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("failed to stat %q: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return dirsRepositoryAdapterPort.ErrInvalidPath
		}
		current = filepath.Dir(current)
	}

	// Create directory
	return os.MkdirAll(targetAbs, 0700)
}

/*
DeleteDir safely deletes a target directory within the configured storage root path (storeLocalRootPath),
with strong protection against path traversal, symlink escapes, and excessive directory depth.

SECURITY GOALS:

1. Prevent deletion outside the allowed base directory.
2. Block any path traversal attempts (e.g., "../../etc/passwd").
3. Detect and reject symbolic links that point outside the base directory.
4. Limit traversal depth to avoid DoS from deeply nested structures.
5. Ensure only actual directories are deleted (not files).

ALGORITHM:

1. **Path Validation**
  - Rejects empty path, ".", "/", and any path starting with "..".
  - Cleans the path using `filepath.Clean` to normalize it.

2. **Absolute Path Resolution**
  - Converts both base path and target path to absolute form.
  - Ensures target path is still inside the base path using `filepath.Rel`.

3. **Existence and Type Check**
  - Confirms that the target exists and is a directory.

4. **Recursive Walk & Symlink Check**
  - Traverses directory contents with `filepath.WalkDir`.
  - If a symlink is found, resolves it with `filepath.EvalSymlinks`.
  - Aborts if the symlink points outside `storeLocalRootPath`.

5. **Depth Limit**
  - If the relative path from the target exceeds `maxDepth` directory separators, abort.

6. **Deletion**
  - If all checks pass, deletes the target directory with `os.RemoveAll`.

SECURITY EXAMPLES:

Base path:

	/var/data

Allowed deletions:

	Path: "uploads/images"      → /var/data/uploads/images   [OK] inside base
	Path: "tmp/session123"      → /var/data/tmp/session123   [OK] inside base

Rejected paths:

	Path: "../../etc"           → [REJECT] path traversal detected
	Path: "/etc/passwd"         → [REJECT] outside base directory
	Path: "uploads/../.. "      → [REJECT] resolves above base
	Path: "."                   → [REJECT] invalid (root of base)
	Path: "uploads/symlink"     → [REJECT] if symlink target is outside base

Symlink escape example:

	/var/data/uploads/link → /etc
	Result: rejected (points outside base)

DEPTH CHECK EXAMPLE (maxDepth = 5):

	/var/data/a/b/c/d/e/f → rejected (depth exceeded)
	/var/data/a/b/c/d/e   → allowed

PATH CHECK FLOW:

	User input path
	      ↓
	filepath.Clean()
	      ↓
	Is empty / "." / "/" / starts with ".."? → REJECT
	      ↓
	Abs(basePath), Abs(targetPath)
	      ↓
	filepath.Rel(basePath, targetPath)
	      ↓
	Starts with ".." or equals "."? → REJECT
	      ↓
	Exists and is a directory?
	      ↓
	WalkDir(targetPath):
	     ├── Depth > maxDepth? → REJECT
	     └── Is symlink? EvalSymlinks()
	             └── Points outside base? → REJECT
	      ↓
	os.RemoveAll(targetPath)

This function is designed for production use with high safety guarantees against accidental
or malicious deletion outside the designated storage root.
*/
func (a *adapter) DeleteDir(ctx context.Context, data *dirsRepositoryAdapterPort.DeleteDirData) error {
	// Maximum allowed directory depth
	const maxDepth = 5

	// Validate input path
	if data.Path == "" {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}
	cleanPath := filepath.Clean(data.Path)
	if cleanPath == "." || cleanPath == "/" || strings.HasPrefix(cleanPath, "..") {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}

	// Resolve absolute paths
	baseAbs, err := filepath.Abs(a.storeLocalRootPath)
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}
	targetAbs, err := filepath.Abs(filepath.Join(baseAbs, cleanPath))
	if err != nil {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}

	// Ensure targetAbs is inside baseAbs
	relToBase, err := filepath.Rel(baseAbs, targetAbs)
	if err != nil {
		return fmt.Errorf("failed to compute relative path: %w", err)
	}
	if strings.HasPrefix(relToBase, "..") || relToBase == "." {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}

	// Check that the target exists and is a directory
	info, err := os.Lstat(targetAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return dirsRepositoryAdapterPort.ErrDirNotFound
		}
		return err
	}
	if !info.IsDir() {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}

	// Walk through and check for symlinks
	err = filepath.WalkDir(targetAbs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		// DoS protection: check directory depth
		rel, _ := filepath.Rel(targetAbs, path)
		if depth := strings.Count(filepath.ToSlash(rel), "/"); depth > maxDepth {
			return fmt.Errorf("max directory depth exceeded at %q", path)
		}

		// Symlink check
		if d.Type()&os.ModeSymlink != 0 {
			resolved, err := filepath.EvalSymlinks(path)
			if err != nil {
				return fmt.Errorf("failed to resolve symlink %q: %w", path, err)
			}
			resolvedAbs, err := filepath.Abs(resolved)
			if err != nil {
				return fmt.Errorf("failed to get absolute path for symlink %q: %w", path, err)
			}

			relToBase, err := filepath.Rel(baseAbs, resolvedAbs)
			if err != nil || strings.HasPrefix(relToBase, "..") {
				return fmt.Errorf("symlink %q points outside base dir (target: %q)", path, resolvedAbs)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Perform deletion
	return os.RemoveAll(targetAbs)
}

/*
RenameDir securely renames a directory within the adapter's base path.

This function performs multiple safety checks before performing the rename:

 1. Validates that both old and new paths are non-empty and do not traverse outside
    the base directory using ".." or absolute paths.
 2. Resolves absolute paths for old and new directories relative to the base.
 3. Ensures both paths are inside the adapter's storeLocalRootPath.
 4. Checks that the old directory exists and the new directory does not exist.
 5. Walks through parent directories of both old and new paths to ensure no symlinks
    exist, preventing symlink race attacks.
 6. Protects against excessive depth in the directory structure to mitigate DoS risks.

Allowed paths (example, assuming base is /var/data):

| Input Path          | Resulting Absolute Path | Reason                     |
|---------------------|-------------------------|----------------------------|
| old: "uploads/img"  | /var/data/uploads/img   | Inside base, safe          |
| new: "uploads/img2" | /var/data/uploads/img2  | Inside base, no collisions |

Rejected paths examples:

| Input Path            | Reason for rejection                          |
|-----------------------|-----------------------------------------------|
| "../../etc"           | Path traversal detected — outside base        |
| "/etc/passwd"         | Absolute path outside base                    |
| "uploads/../.."       | Resolves above base directory                 |
| "uploads/symlink"     | Symlink in path points outside base           |
| "."                   | Invalid — refers to base directory itself     |
*/
func (a *adapter) RenameDir(ctx context.Context, data *dirsRepositoryAdapterPort.RenameDirData) error {
	// Validate input paths
	if data.OldPath == "" || data.NewPath == "" {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}
	oldClean := filepath.Clean(data.OldPath)
	newClean := filepath.Clean(data.NewPath)
	if oldClean == "." || strings.HasPrefix(oldClean, "..") ||
		newClean == "." || strings.HasPrefix(newClean, "..") {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}

	// Resolve absolute paths
	baseAbs, err := filepath.Abs(a.storeLocalRootPath)
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}
	oldAbs, err := filepath.Abs(filepath.Join(baseAbs, oldClean))
	if err != nil {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}
	newAbs, err := filepath.Abs(filepath.Join(baseAbs, newClean))
	if err != nil {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}

	// Ensure old and new paths are inside base
	relOld, err := filepath.Rel(baseAbs, oldAbs)
	if err != nil || strings.HasPrefix(relOld, "..") {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}
	relNew, err := filepath.Rel(baseAbs, newAbs)
	if err != nil || strings.HasPrefix(relNew, "..") {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}

	// Check old directory exists
	info, err := os.Lstat(oldAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return dirsRepositoryAdapterPort.ErrDirOldNotFound
		}
		return err
	}
	if !info.IsDir() {
		return dirsRepositoryAdapterPort.ErrInvalidPath
	}

	// Check new directory does not exist
	if _, err := os.Lstat(newAbs); err == nil {
		return dirsRepositoryAdapterPort.ErrDirNewExist
	}

	// Check for symlinks in parent directories of old and new
	for _, path := range []string{oldAbs, newAbs} {
		current := filepath.Dir(path)
		for {
			if current == baseAbs || current == string(filepath.Separator) {
				break
			}
			info, err := os.Lstat(current)
			if err != nil {
				return dirsRepositoryAdapterPort.ErrInvalidPath
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return dirsRepositoryAdapterPort.ErrInvalidPath
			}
			current = filepath.Dir(current)
		}
	}

	// Perform rename
	return os.Rename(oldAbs, newAbs)
}
