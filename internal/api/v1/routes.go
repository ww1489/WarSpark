package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/ww1489/WarSpark/docs"
	appconfig "github.com/ww1489/WarSpark/internal/config"
	"github.com/ww1489/WarSpark/internal/controller"
	infracoc "github.com/ww1489/WarSpark/internal/infra/coc"
	infraredis "github.com/ww1489/WarSpark/internal/infra/redis"
	authmw "github.com/ww1489/WarSpark/internal/middleware/auth"
	"github.com/ww1489/WarSpark/internal/repository"
	"github.com/ww1489/WarSpark/internal/service"
	cocapi "github.com/ww1489/WarSpark/pkg/cocapi"
)

func SetupRoutes(router *gin.Engine, runtimeConfig appconfig.RuntimeConfig) {
	healthController := controller.NewHealthController(runtimeConfig.MySQL, runtimeConfig.Redis)
	layoutRepository := repository.NewLayoutRepository(runtimeConfig.MySQL)
	layoutService := service.NewLayoutService(layoutRepository)
	layoutController := controller.NewLayoutController(layoutService)
	adminLayoutController := controller.NewAdminLayoutController(layoutService)
	imageSearchRepository := repository.NewImageSearchRepository(runtimeConfig.MySQL)
	imageStorage := service.NewLocalImageStorage("data/uploads", "/uploads")
	imageSearchService := service.NewImageSearchService(imageSearchRepository, imageStorage)
	imageSearchUploadLimiter := infraredis.NewImageSearchUploadRateLimiter(runtimeConfig.Redis)
	imageSearchController := controller.NewImageSearchController(imageSearchService, controller.ImageSearchControllerOptions{
		UploadLimiter: imageSearchUploadLimiter,
	})
	warRepository := repository.NewWarRepository(runtimeConfig.MySQL)
	warAPIClient := infracoc.New(cocapi.Config{
		BaseURL:  runtimeConfig.Config.CoC.BaseURL,
		APIToken: runtimeConfig.Config.CoC.APIToken,
		Timeout:  runtimeConfig.Config.CoC.Timeout,
	})
	warCache := infraredis.NewWarCache(runtimeConfig.Redis)
	warService := service.NewWarService(warAPIClient, warRepository, service.WarServiceOptions{
		Cache:              warCache,
		CurrentWarCacheTTL: runtimeConfig.Config.CoC.CurrentWarCacheTTL,
		CWLGroupCacheTTL:   runtimeConfig.Config.CoC.CWLGroupCacheTTL,
	})
	warController := controller.NewWarController(warService)

	router.GET("/health", healthController.Check)
	router.Static("/uploads", "data/uploads")
	if runtimeConfig.Config.Docs.SwaggerEnabled {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	api := router.Group("/api/v1")
	{
		api.GET("/health", healthController.Check)
		api.GET("/ready", healthController.Ready)
		api.GET("/layouts", layoutController.List)
		api.GET("/layouts/:layout_id", layoutController.Get)
		api.GET("/layouts/:layout_id/videos", layoutController.ListVideos)
		api.POST("/image-search/jobs", imageSearchController.CreateJob)
		api.GET("/image-search/jobs/:job_id", imageSearchController.GetJob)
		api.GET("/image-search/jobs/:job_id/results", imageSearchController.GetResults)
		api.POST("/image-search/jobs/:job_id/retry", imageSearchController.RetryJob)
		api.GET("/war/current", warController.GetCurrent)
		api.GET("/war/cwl", warController.GetCWL)
		api.GET("/war/snapshots/:war_snapshot_id/members", warController.ListMembers)

		admin := api.Group("/admin", authmw.Required(runtimeConfig.TokenManager))
		{
			admin.GET("/review-queue", adminLayoutController.ListReviewQueue)
			admin.GET("/audit-logs", adminLayoutController.ListAuditLogs)
			admin.POST("/layouts", adminLayoutController.CreateDraft)
			admin.POST("/layouts/:layout_id/links", adminLayoutController.AddLink)
			admin.POST("/layouts/:layout_id/video-matches", adminLayoutController.AddVideoMatch)
			admin.PATCH("/review/:resource_type/:resource_id", adminLayoutController.UpdateReviewStatus)
			admin.PATCH("/layout-links/:link_id", adminLayoutController.UpdateLink)
		}

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
