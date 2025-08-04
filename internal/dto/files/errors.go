package dto

import (
	"github.com/flash-go/sdk/errors"
)

var (
	ErrDirInvalidPath    = errors.New(errors.ErrBadRequest, "invalid_path")
	ErrDirInvalidOldPath = errors.New(errors.ErrBadRequest, "invalid_old_path")
	ErrDirInvalidNewPath = errors.New(errors.ErrBadRequest, "invalid_new_path")
)
