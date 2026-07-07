package plugins

import (
	contentaudit "cybros/plugins/content_audit"
	memberdump "cybros/plugins/member_dump"

	"github.com/gotd/td/tg"
)

func New() []Plugin {
	return []Plugin{
		contentaudit.New(),
		memberdump.New(),
	}
}

func SetAPI(pluginList []Plugin, api *tg.Client) {
	for _, plugin := range pluginList {
		apiPlugin, ok := plugin.(APIPlugin)
		if ok {
			apiPlugin.SetAPI(api)
		}
	}
}
