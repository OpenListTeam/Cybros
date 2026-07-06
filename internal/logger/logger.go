package logger

import (
	"os"
	"strings"

	"cybros/consts"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func Init() {
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "info"
	}
	level, err := logrus.ParseLevel(strings.ToLower(levelStr))
	if err != nil {
		logrus.Fatalf(consts.ErrorInvalidLogLevel, levelStr)
	}

	Log = logrus.New()
	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors:               true,
		EnvironmentOverrideColors: true,
		TimestampFormat:           "2006-01-02T15:04:05.000000",
		FullTimestamp:             true,
	})

	Log.SetLevel(level)
	Log.SetOutput(os.Stdout)
}
