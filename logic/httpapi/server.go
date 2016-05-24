package httpapi

import (
	"net/http"

	"log"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
)

// Start run the http server.
func Start(config *ServerConfig) {
	setupKeys(config)

	e := echo.New()
	e.SetDebug(config.Debug)
	e.Use(middleware.Logger())
	e.GET("/", test)

	setupAppAPI(e)
	setupUserAPI(e)

	log.Println("http listening:", config.Addr)
	e.Run(standard.New(config.Addr))
}

func setupAppAPI(e *echo.Echo) {
	gApp := e.Group("/xim/app")
	c := middleware.DefaultJWTAuthConfig
	c.ContextKey = "appToken"
	c.SigningKey = appKey
	c.Extractor = jwtFromHeaderOrQueryParam
	gApp.Use(middleware.JWTAuthWithConfig(c))
	gApp.Use(fetchApp)

	gApp.Get("/msg/:channel/last_id", getChannelLastID)
}

func setupUserAPI(e *echo.Echo) {
	gUserXim := e.Group("/user/xim")
	gUserXim.Use(middleware.JWTAuth(userKey))
	gUserXim.GET("", test)
}

func test(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func jwtFromHeaderOrQueryParam(c echo.Context) (string, error) {
	token, err := middleware.JWTFromHeader(c)
	if err != nil {
		token = c.FormValue("jwt")
	}
	if token != "" {
		return token, nil
	}
	return "", err
}
