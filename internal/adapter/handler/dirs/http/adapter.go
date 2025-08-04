package adapter

import (
	dto "github.com/flash-go/files-service/internal/dto/dirs"
	httpDirsHandlerAdapterPort "github.com/flash-go/files-service/internal/port/adapter/handler/dirs/http"
	dirsServicePort "github.com/flash-go/files-service/internal/port/service/dirs"
	"github.com/flash-go/flash/http/server"
	"github.com/flash-go/sdk/errors"
)

type Config struct {
	DirsService dirsServicePort.Interface
}

func New(config *Config) httpDirsHandlerAdapterPort.Interface {
	return &adapter{
		config.DirsService,
	}
}

type adapter struct {
	dirsService dirsServicePort.Interface
}

// @Summary Create dir (admin)
// @Tags dirs
// @Security BearerAuth
// @Accept json
// @Produce plain
// @Param request body dto.AdminCreateDirRequest true "Create dir (admin)"
// @Success 201
// @Failure 400 {string} string "Possible error codes: bad_request, bad_request:invalid_path, bad_request:dir_exist"
// @Router /admin/dirs [post]
func (a *adapter) AdminCreateDir(ctx server.ReqCtx) {
	// Parse request json body
	var request dto.AdminCreateDirRequest
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
	data := dirsServicePort.CreateDirData(request)

	// Create dir
	if err := a.dirsService.CreateDir(
		ctx.Context(),
		&data,
	); err != nil {
		ctx.WriteErrorResponse(err)
		return
	}

	// Write success response
	ctx.WriteResponse(201, nil)
}

// @Summary Delete dir (admin)
// @Tags dirs
// @Security BearerAuth
// @Accept json
// @Produce plain
// @Param request body dto.AdminDeleteDirRequest true "Delete dir (admin)"
// @Success 200
// @Failure 400 {string} string "Possible error codes: bad_request, bad_request:invalid_path, bad_request:dir_not_found"
// @Router /admin/dirs [delete]
func (a *adapter) AdminDeleteDir(ctx server.ReqCtx) {
	// Parse request json body
	var request dto.AdminDeleteDirRequest
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
	data := dirsServicePort.DeleteDirData(request)

	// Delete dir
	if err := a.dirsService.DeleteDir(
		ctx.Context(),
		&data,
	); err != nil {
		ctx.WriteErrorResponse(err)
		return
	}

	// Write success response
	ctx.WriteResponse(200, nil)
}

// @Summary Rename dir (admin)
// @Tags dirs
// @Security BearerAuth
// @Accept json
// @Produce plain
// @Param request body dto.AdminRenameDirRequest true "Rename dir (admin)"
// @Success 200
// @Failure 400 {string} string "Possible error codes: bad_request, bad_request:invalid_old_path, bad_request:invalid_new_path, bad_request:old_dir_not_found, bad_request:new_dir_exist"
// @Router /admin/dirs [patch]
func (a *adapter) AdminRenameDir(ctx server.ReqCtx) {
	// Parse request json body
	var request dto.AdminRenameDirRequest
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
	data := dirsServicePort.RenameDirData(request)

	// Rename dir
	if err := a.dirsService.RenameDir(
		ctx.Context(),
		&data,
	); err != nil {
		ctx.WriteErrorResponse(err)
		return
	}

	// Write success response
	ctx.WriteResponse(200, nil)
}
