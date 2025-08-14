package port

import "github.com/flash-go/sdk/errors"

var (
	ErrInvalidPath     = errors.New(errors.ErrBadRequest, "invalid_path")
	ErrInvalidFile     = errors.New(errors.ErrBadRequest, "invalid_file")
	ErrFileExist       = errors.New(errors.ErrBadRequest, "file_exist")
	ErrDirNotFound     = errors.New(errors.ErrBadRequest, "dir_not_found")
	ErrFileNotFound    = errors.New(errors.ErrBadRequest, "file_not_found")
	ErrFileOldNotFound = errors.New(errors.ErrBadRequest, "old_file_not_found")
	ErrFileNewExist    = errors.New(errors.ErrBadRequest, "new_file_exist")
)
