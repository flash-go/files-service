package factory

import (
	"encoding/json"
	"time"

	"github.com/flash-go/files-service/internal/domain/entity"
)

func NewClient(data ClientData) *entity.Client {
	return &entity.Client{
		ClientType: data.ClientType,
		Metadata:   data.Metadata,
		Created:    time.Unix(0, data.Created.UnixNano()),
	}
}

type ClientData struct {
	ClientType string
	Metadata   json.RawMessage
	Created    time.Time
}
