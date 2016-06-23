package logger

import (
	"github.com/go-ozzo/ozzo-log"
)

// root logger.
var (
	Logger = log.NewLogger()
)

func init() {
	t1 := log.NewConsoleTarget()
	t1.MaxLevel = log.LevelDebug
	Logger.Targets = append(Logger.Targets, t1)
	Logger.Open()
}
