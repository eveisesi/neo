package neo

import (
	"time"

	"github.com/volatiletech/null"
)

type ServerStatus struct {
	Players       int       `json:"players"`
	ServerVersion string    `json:"server_version"`
	StartTime     time.Time `json:"start_time"`
	VIP           null.Bool `json:"vip"`
}
