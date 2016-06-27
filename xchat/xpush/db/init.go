package token

import (
	"xim/xchat/logic/logger"

	ol "github.com/go-ozzo/ozzo-log"
)

// variables
var (
	l *ol.Logger
)

func init() {
	l = logger.Logger.GetLogger("token")
}
