package dto

type AdminCreateDirRequest struct {
	Path string `json:"path"`
}

func (r *AdminCreateDirRequest) Validate() error {
	if err := r.ValidatePath(); err != nil {
		return err
	}
	return nil
}

func (r *AdminCreateDirRequest) ValidatePath() error {
	if r.Path == "" {
		return ErrDirInvalidPath
	}
	return nil
}

type AdminDeleteDirRequest struct {
	Path string `json:"path"`
}

func (r *AdminDeleteDirRequest) Validate() error {
	if err := r.ValidatePath(); err != nil {
		return err
	}
	return nil
}

func (r *AdminDeleteDirRequest) ValidatePath() error {
	if r.Path == "" {
		return ErrDirInvalidPath
	}
	return nil
}

type AdminRenameDirRequest struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}

func (r *AdminRenameDirRequest) Validate() error {
	if err := r.ValidateOldPath(); err != nil {
		return err
	}
	if err := r.ValidateNewPath(); err != nil {
		return err
	}
	return nil
}

func (r *AdminRenameDirRequest) ValidateOldPath() error {
	if r.OldPath == "" {
		return ErrDirInvalidOldPath
	}
	return nil
}

func (r *AdminRenameDirRequest) ValidateNewPath() error {
	if r.NewPath == "" {
		return ErrDirInvalidNewPath
	}
	return nil
}
