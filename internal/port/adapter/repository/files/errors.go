package port

import "github.com/flash-go/sdk/errors"

var (
	ErrFileExist       = errors.New(errors.ErrBadRequest, "file_exist")
	ErrDirNotFound     = errors.New(errors.ErrBadRequest, "dir_not_found")
	ErrFileNotFound    = errors.New(errors.ErrBadRequest, "file_not_found")
	ErrFileOldNotFound = errors.New(errors.ErrBadRequest, "old_file_not_found")
	ErrFileNewExist    = errors.New(errors.ErrBadRequest, "new_file_exist")
)
