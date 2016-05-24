package httpapi

import (
	"log"
	"net/http"

	"xim/commons/db"
	"xim/commons/msgdb"

	"github.com/labstack/echo"
)

func getChannelLastID(c echo.Context) error {
	app := c.Get("app").(string)
	channel := c.Param("channel")
	if db.AppChannelExists(app, channel) {
		msgStore := msgdb.GetMsgStore()
		defer msgStore.Close()
		id, err := msgStore.LastID(channel)
		if err != nil {
			log.Println("get last msg id err:", channel, err)
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"ok": true,
			"id": id,
		})
	}
	return c.JSON(http.StatusBadRequest, map[string]interface{}{
		"ok": false,
	})
}
