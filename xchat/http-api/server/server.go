package server

import (
	"net/http"
	"xim/utils/nanorpc"
	"xim/xchat/logic/logger"

	ol "github.com/go-ozzo/ozzo-log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
)

var (
	l *ol.Logger
)

var (
	config     *Config
	xchatLogic *nanorpc.Client
)

func init() {
	l = logger.Logger.GetLogger("server")
}

// Setup initialze mid.
func Setup(config *Config) {
	xchatLogic = nanorpc.NewClient(config.LogicRPCAddr, config.RPCCallTimeout)
}

// Start run the http server.
func Start(c *Config) {
	config = c
	xchatLogic = nanorpc.NewClient(config.LogicRPCAddr, config.RPCCallTimeout)

	e := echo.New()
	e.SetDebug(config.Debug)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	if config.Testing {
		e.GET("/test/", test)
	}

	setup(e)

	l.Info("http listening: %s", config.Addr)
	e.Run(standard.New(config.Addr))
}

func setup(e *echo.Echo) {
	gXChatAPI := e.Group("/xchat/api")

	gXChatAPI.Use(JWT("token", config.Key))
	gXChatAPI.Use(RequireIsAdminUser)
	gXChatAPI.GET("/test/", test)

	gXChatAPI.Post("/user/msg/send/", sendMsg)
}

func test(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
