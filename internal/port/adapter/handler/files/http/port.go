package port

import (
	"github.com/flash-go/flash/http/server"
)

type Interface interface {
	AdminCreateFile(ctx server.ReqCtx)
	AdminListFiles(ctx server.ReqCtx)
	AdminDeleteFile(ctx server.ReqCtx)
	AdminRenameFile(ctx server.ReqCtx)
}
