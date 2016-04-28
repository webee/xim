package httpapi

import (
	"encoding/base64"
	"net/http"
	"time"

	"golang.org/x/crypto/scrypt"

	"xim/logic/db"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
)

func genPassword(raw string) string {
	dk, err := scrypt.Key([]byte(raw), salt, 16384, 8, 1, 32)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(dk)
}

func appAuth(app, password string) (*db.App, bool) {
	ximApp, err := db.GetApp(app)
	if err != nil {
		return nil, false
	}
	if !ximApp.Password.Valid {
		return nil, false
	}
	return ximApp, genPassword(password) == ximApp.Password.String
}

// appNewToken use app and password to get an app api access token.
func appNewToken(c echo.Context) error {
	app := c.Request().Header().Get("Xim-App")
	password := c.Request().Header().Get("Xim-App-Password")

	ximApp, ok := appAuth(app, password)
	if !ok {
		c.Response().Header().Set("Xim-App-Auth", "Restricted")
		return echo.ErrUnauthorized
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	token.Claims["app"] = ximApp.App
	token.Claims["exp"] = time.Now().Add(6 * time.Hour).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString(appKey)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"ok":    true,
		"token": t,
	})
}

// newUserToken is an app api to generate a user access token.
func newUserToken(c echo.Context) (err error) {
	appToken := c.Get("app").(*jwt.Token)
	app := appToken.Claims["app"].(string)
	user := c.QueryParam("user")
	expire := c.QueryParam("expire")

	if user == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"ok":  false,
			"err": "bad user",
		})
	}
	exp := 10 * time.Minute
	if expire != "" {
		exp, err = time.ParseDuration(expire)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"ok":  false,
				"err": "bad expire",
			})
		}
	}

	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	token.Claims["app"] = app
	token.Claims["user"] = user
	token.Claims["exp"] = time.Now().Add(exp).Unix()

	// Generate encoded token and send it as response.
	t, err := token.SignedString(userKey)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"ok":    true,
		"token": t,
	})
}
