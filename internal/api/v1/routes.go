package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/ww1489/WarSpark/docs"
	appconfig "github.com/ww1489/WarSpark/internal/config"
	"github.com/ww1489/WarSpark/internal/controller"
)

func SetupRoutes(router *gin.Engine, runtimeConfig appconfig.RuntimeConfig) {
	healthController := controller.NewHealthController(runtimeConfig.MySQL, runtimeConfig.Redis)

	router.GET("/health", healthController.Check)
	if runtimeConfig.Config.Docs.SwaggerEnabled {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := router.Group("/api/v1")
	{
		api.GET("/health", healthController.Check)
		api.GET("/ready", healthController.Ready)

		api.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "ok",
				"data": gin.H{
					"name":         appconfig.AppName,
					"display_name": appconfig.AppDisplayName,
					"version":      appconfig.AppVersion,
				},
			})
		})
	}
}
