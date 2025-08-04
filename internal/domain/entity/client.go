package entity

import (
	"encoding/json"
	"time"
)

type Client struct {
	Id         uint
	ClientType string
	Metadata   json.RawMessage
	Created    time.Time
}
