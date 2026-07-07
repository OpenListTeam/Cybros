package memberdump

import (
	"sync"

	"github.com/gotd/td/tg"
)

type MemberDump struct {
	api *tg.Client

	mu      sync.Mutex
	running bool
}

func New() *MemberDump {
	return &MemberDump{}
}

func (p *MemberDump) SetAPI(api *tg.Client) {
	p.api = api
}
