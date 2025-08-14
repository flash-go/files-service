package port

import "github.com/flash-go/sdk/errors"

var (
	ErrInvalidPath    = errors.New(errors.ErrBadRequest, "invalid_path")
	ErrDirExist       = errors.New(errors.ErrBadRequest, "dir_exist")
	ErrDirNotFound    = errors.New(errors.ErrBadRequest, "dir_not_found")
	ErrDirOldNotFound = errors.New(errors.ErrBadRequest, "old_dir_not_found")
	ErrDirNewExist    = errors.New(errors.ErrBadRequest, "new_dir_exist")
)
