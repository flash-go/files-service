package dto

type AdminCreateFileRequest struct {
	Path string `json:"path"`
}

type AdminListFilesRequest struct {
	Path string `json:"path"`
}

type AdminDeleteFileRequest struct {
	Path string `json:"path"`
}

func (r *AdminDeleteFileRequest) Validate() error {
	if err := r.ValidatePath(); err != nil {
		return err
	}
	return nil
}

func (r *AdminDeleteFileRequest) ValidatePath() error {
	if r.Path == "" {
		return ErrDirInvalidPath
	}
	return nil
}

type AdminRenameFileRequest struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}

func (r *AdminRenameFileRequest) Validate() error {
	if err := r.ValidateOldPath(); err != nil {
		return err
	}
	if err := r.ValidateNewPath(); err != nil {
		return err
	}
	return nil
}

func (r *AdminRenameFileRequest) ValidateOldPath() error {
	if r.OldPath == "" {
		return ErrDirInvalidOldPath
	}
	return nil
}

func (r *AdminRenameFileRequest) ValidateNewPath() error {
	if r.NewPath == "" {
		return ErrDirInvalidNewPath
	}
	return nil
}
