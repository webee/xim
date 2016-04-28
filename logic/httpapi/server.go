package httpapi

import (
  "net/http"
	"golang.org/x/crypto/scrypt"
  "github.com/labstack/echo"
  "github.com/labstack/echo/engine/standard"
)

func init() {
	scrypt.Key(password []byte, salt []byte, N int, r int, p int, keyLen int)
}


func Start() {
  e := echo.New()
  e.GET("/", test)

  go e.Run(standard.New(":1323"))
}


func test(c echo.Context) error {
  return c.String(http.StatusOK, "OK")
}
