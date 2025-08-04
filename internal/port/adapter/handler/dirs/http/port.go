package port

import (
	"github.com/flash-go/flash/http/server"
)

type Interface interface {
	AdminCreateDir(ctx server.ReqCtx)
	AdminDeleteDir(ctx server.ReqCtx)
	AdminRenameDir(ctx server.ReqCtx)
}
