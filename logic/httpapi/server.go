package httpapi

import (
	"net/http"

	"log"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
)

func init() {
	//scrypt.Key(password []byte, salt []byte, N int, r int, p int, keyLen int)
}

// Start run the http server.
func Start(config *ServerConfig) {
	setupKeys(config)

	e := echo.New()
	e.SetDebug(config.Debug)
	e.Use(middleware.Logger())
	e.GET("/", test)

	e.GET("app.new_token", appNewToken)

	gAppXim := e.Group("/app/xim")
	c := middleware.DefaultJWTAuthConfig
	c.ContextKey = "app"
	c.SigningKey = appKey
	gAppXim.Use(middleware.JWTAuthWithConfig(c))
	gAppXim.GET(".new_user_token", newUserToken)

	gUserXim := e.Group("/user/xim")
	gUserXim.Use(middleware.JWTAuth(userKey))
	gUserXim.GET("", test)

	log.Println("http listening:", config.Addr)
	e.Run(standard.New(config.Addr))
}

func test(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
