package main

import (
	"net/http"

	"github.com/prayer-time/client/redis"
	"github.com/prayer-time/client/waktusholat"
	"github.com/prayer-time/config"
	"github.com/prayer-time/handler"
	"github.com/prayer-time/service/prayerTime"
	"github.com/prayer-time/util"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Get()

	router := initRouter(cfg)

	util.RunServerGracefully(cfg.PORT, cfg.TimeOut, router)
}

func initRouter(cfg config.Config) *gin.Engine {
	// init service
	waktuSholatSvc := waktusholat.NewService(cfg.WaktuSholatHost, cfg.ApiPrayZoneHost, cfg.DebugLog)
	redisSvc := redis.NewService(redis.RedisConfig{
		TlsUrl:    cfg.RedisTlsUrl,
		Host:      cfg.RedisHost,
		Port:      cfg.RedisPort,
		Password:  cfg.RedisPassword,
		Timeout:   cfg.RedisTimeout,
		MaxIdle:   cfg.RedisMaxIdle,
		MaxActive: cfg.RedisMaxActive,
	})
	prayerTimeSvc := prayerTime.NewService(waktuSholatSvc, redisSvc, cfg.Host, cfg.ExpiredKey)

	// init handler
	prayerTimeHandler := handler.NewHandler(prayerTimeSvc)

	router := gin.Default()

	router.Use(util.CORSMiddleware())

	router.LoadHTMLGlob("views/pages/*")
	router.StaticFile("/favicon.ico", "./views/assets/images/favicon.ico")
	router.Static("/css", "./views/assets/css")
	router.Static("/js", "./views/assets/js")
	router.Static("/images", "./views/assets/images")
	router.Static("/fonts", "./views/assets/fonts")

	router.GET("", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Prayer Time For Calendar",
		})
	})

	route := router.Group("/prayer-time")
	{
		route.POST("/get-key", prayerTimeHandler.GetKeyPrayerTime)
		route.GET("/get", prayerTimeHandler.GetDataPrayerTime)
		route.GET("/get-city", prayerTimeHandler.GetCityByName)
	}

	return router
}
