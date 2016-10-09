package server

import (
	"net/http"
	"xim/utils/nanorpc"
	"xim/xchat/logic/logger"
	"xim/xchat/xchat-http-client"

	ol "github.com/go-ozzo/ozzo-log"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
)

// constants.
const (
	NsContextKey    = "ns"
	TokenContextKey = "token"
)

var (
	l *ol.Logger
)

var (
	config          *Config
	xchatLogic      *nanorpc.Client
	xchatHTTPClient *xchathttpclient.XChatHTTPClient
)

func init() {
	l = logger.Logger.GetLogger("server")
}

// Start run the http server.
func Start(c *Config) {
	config = c
	xchatLogic = nanorpc.NewClient(config.LogicRPCAddr, config.RPCCallTimeout)
	xchatHTTPClient = xchathttpclient.NewXChatHTTPClient(config.Key, config.XChatHostURL)

	e := echo.New()
	e.SetDebug(config.Debug)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	if config.Testing {
		e.GET("/test/", test)
	}

	setup(e)
	setupXRTC(e)

	l.Info("http listening: %s", config.Addr)
	e.Run(standard.New(config.Addr))
}

func setup(e *echo.Echo) {
	gXChatAPI := e.Group("/xchat/api")

	gXChatAPI.Use(JWT(NsContextKey, TokenContextKey, config.Keys))
	gXChatAPI.Use(RequireIsAdminUser(TokenContextKey))
	gXChatAPI.GET("/test/", test)

	gXChatAPI.Post("/user/msg/send/", sendMsg)
	gXChatAPI.Post("/user/msg/send/unique/chat/", sendUniqueChatMsg)
	gXChatAPI.Post("/user/notify/send/", sendUserNotify)
}

func test(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func setupXRTC(e *echo.Echo) {
	gXRTCAPI := e.Group("/xrtc/api")

	gXRTCAPI.Use(middleware.CORS())
	gXRTCAPI.GET("/iceconfig", fetchIceConfig)
}
