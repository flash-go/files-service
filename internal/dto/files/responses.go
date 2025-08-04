package dto

type FileResponse struct {
	Name     string  `json:"name"`
	IsDir    bool    `json:"is_dir"`
	Size     *int64  `json:"size"`
	MimeType *string `json:"mime_type"`
}
