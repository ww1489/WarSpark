package cors

import (
	gincors "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	appconfig "github.com/ww1489/WarSpark/internal/config"
)

func New(cfg appconfig.CORSConfig) gin.HandlerFunc {
	return gincors.New(gincors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     cfg.AllowedMethods,
		AllowHeaders:     cfg.AllowedHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	})
}
