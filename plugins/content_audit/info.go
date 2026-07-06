package contentaudit

import "github.com/gotd/td/tg"

type ContentAudit struct {
	api *tg.Client
}

func New() *ContentAudit {
	return &ContentAudit{}
}

func (p *ContentAudit) SetAPI(api *tg.Client) {
	p.api = api
}
