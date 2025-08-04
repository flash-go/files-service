package adapter

import (
	"encoding/json"

	dto "github.com/flash-go/files-service/internal/dto/files"
	httpFilesHandlerAdapterPort "github.com/flash-go/files-service/internal/port/adapter/handler/files/http"
	filesServicePort "github.com/flash-go/files-service/internal/port/service/files"
	"github.com/flash-go/flash/http/server"
	"github.com/flash-go/sdk/errors"
)

type Config struct {
	FilesService filesServicePort.Interface
}

func New(config *Config) httpFilesHandlerAdapterPort.Interface {
	return &adapter{
		config.FilesService,
	}
}

type adapter struct {
	filesService filesServicePort.Interface
}

// @Summary Create file (admin)
// @Tags files
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce plain
// @Param file formData file true "File to upload"
// @Param meta formData string true "Metadata"
// @Success 201
// @Failure 400 {string} string "Possible error codes: bad_request, bad_request:dir_not_found, bad_request:file_exist"
// @Router /admin/files [post]
func (a *adapter) AdminCreateFile(ctx server.ReqCtx) {
	// Get request file
	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.WriteErrorResponse(err)
		return
	}

	// Parse request json metadata
	var request dto.AdminCreateFileRequest
	if err := json.Unmarshal(
		ctx.FormValue("meta"),
		&request,
	); err != nil {
		ctx.WriteErrorResponse(errors.ErrBadRequest)
		return
	}

	// Create file
	if err := a.filesService.CreateFile(
		ctx.Context(),
		&filesServicePort.CreateFileData{
			Path: request.Path,
			File: file,
		},
	); err != nil {
		ctx.WriteErrorResponse(err)
		return
	}

	// Write success response
	ctx.WriteResponse(201, nil)
}

// @Summary List files (admin)
// @Tags files
// @Security BearerAuth
// @Accept json
// @Produce json,plain
// @Param request body dto.AdminListFilesRequest true "List files (admin)"
// @Success 200 {array} dto.FileResponse
// @Failure 400 {string} string "Possible error codes: bad_request, bad_request:invalid_path"
// @Router /admin/files/list [post]
func (a *adapter) AdminListFiles(ctx server.ReqCtx) {
	// Parse request json body
	var request dto.AdminListFilesRequest
	if err := ctx.ReadJson(&request); err != nil {
		ctx.WriteErrorResponse(errors.ErrBadRequest)
		return
	}

	// Create data
	data := filesServicePort.GetFilesData(request)

	// Get files
	files, err := a.filesService.GetFiles(
		ctx.Context(),
		&data,
	)
	if err != nil {
		ctx.WriteErrorResponse(err)
		return
	}

	// Build response
	response := make([]dto.FileResponse, len(*files))
	for i, file := range *files {
		response[i] = dto.FileResponse(file)
	}

	// Write success response
	ctx.WriteResponse(200, response)
}

// @Summary Delete file (admin)
// @Tags files
// @Security BearerAuth
// @Accept json
// @Produce plain
// @Param request body dto.AdminDeleteFileRequest true "Delete file (admin)"
// @Success 200
// @Failure 400 {string} string "Possible error codes: bad_request, bad_request:invalid_path, bad_request:file_not_found"
// @Router /admin/files [delete]
func (a *adapter) AdminDeleteFile(ctx server.ReqCtx) {
	// Parse request json body
	var request dto.AdminDeleteFileRequest
	if err := ctx.ReadJson(&request); err != nil {
		ctx.WriteErrorResponse(errors.ErrBadRequest)
		return
	}

	// Validate request
	if err := request.Validate(); err != nil {
		ctx.WriteErrorResponse(err)
		return
	}

	// Create data
	data := filesServicePort.DeleteFileData(request)

	// Delete file
	if err := a.filesService.DeleteFile(
		ctx.Context(),
		&data,
	); err != nil {
		ctx.WriteErrorResponse(err)
		return
	}

	// Write success response
	ctx.WriteResponse(200, nil)
}

// @Summary Rename file (admin)
// @Tags files
// @Security BearerAuth
// @Accept json
// @Produce plain
// @Param request body dto.AdminRenameFileRequest true "Rename file (admin)"
// @Success 200
// @Failure 400 {string} string "Possible error codes: bad_request, bad_request:invalid_old_path, bad_request:invalid_new_path, bad_request:old_file_not_found, bad_request:new_file_exist"
// @Router /admin/files [patch]
func (a *adapter) AdminRenameFile(ctx server.ReqCtx) {
	// Parse request json body
	var request dto.AdminRenameFileRequest
	if err := ctx.ReadJson(&request); err != nil {
		ctx.WriteErrorResponse(errors.ErrBadRequest)
		return
	}

	// Validate request
	if err := request.Validate(); err != nil {
		ctx.WriteErrorResponse(err)
		return
	}

	// Create data
	data := filesServicePort.RenameFileData(request)

	// Rename file
	if err := a.filesService.RenameFile(
		ctx.Context(),
		&data,
	); err != nil {
		ctx.WriteErrorResponse(err)
		return
	}

	// Write success response
	ctx.WriteResponse(200, nil)
}
