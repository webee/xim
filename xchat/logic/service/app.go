package service

import (
	"time"

	"xim/xchat/logic/db"

	"github.com/patrickmn/go-cache"
)

var (
	appsCache = cache.New(15*time.Minute, 3*time.Minute)
)

func getAppInfo(appID string) *db.App {
	value, ok := appsCache.Get(appID)
	if ok {
		return value.(*db.App)
	}
	l.Debug("loading app: %s", appID)
	app, err := db.GetApp(appID)
	if err != nil {
		l.Error("load app error: %+v", err)
		return nil
	}
	l.Debug("loaded app: %+v", app)
	appsCache.Set(appID, app, cache.DefaultExpiration)
	return app
}
