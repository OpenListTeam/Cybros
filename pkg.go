package cybros

import (
	"github.com/gotd/td/tg"
	"github.com/sirupsen/logrus"
)

var (
	_ = logrus.New()
	_ = tg.NewClient(nil)
)
