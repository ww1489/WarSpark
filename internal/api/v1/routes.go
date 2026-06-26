package v1

import (
	"net/http"
	"time"

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

func SetupRoutes(router *gin.Engine, runtimeConfig appconfig.RuntimeConfig, imageSearchService *service.ImageSearchService) {
	healthController := controller.NewHealthController(runtimeConfig.MySQL, runtimeConfig.Redis)
	layoutRepository := repository.NewLayoutRepository(runtimeConfig.MySQL)
	layoutService := service.NewLayoutService(layoutRepository)
	layoutController := controller.NewLayoutController(layoutService)
	adminLayoutController := controller.NewAdminLayoutController(layoutService)
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
	cocapiClient := cocapi.New(cocapi.Config{
		BaseURL:  runtimeConfig.Config.CoC.BaseURL,
		APIToken: runtimeConfig.Config.CoC.APIToken,
		Timeout:  runtimeConfig.Config.CoC.Timeout,
	})
	warCache := infraredis.NewWarCache(runtimeConfig.Redis, runtimeConfig.Config.CoC.CurrentWarCacheTTL, runtimeConfig.Config.CoC.CWLGroupCacheTTL, 5*time.Minute, 5*time.Minute)
	warService := service.NewWarService(warAPIClient, warRepository, service.WarServiceOptions{
		Cache:        warCache,
		CocapiClient: cocapiClient,
	})
	warController := controller.NewWarController(warService)
	clanCache := infraredis.NewClanCache(runtimeConfig.Redis, 5*time.Minute)
	clanService := service.NewClanService(cocapiClient, clanCache)
	clanController := controller.NewClanController(clanService)
	playerService := service.NewPlayerService(cocapiClient, clanCache)
	playerController := controller.NewPlayerController(playerService)

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
		api.GET("/war/cwl/wars/:war_tag", warController.GetCWLWar)
		api.GET("/war/snapshots/:war_snapshot_id/members", warController.ListMembers)
		api.GET("/clans/:tag/war-log", warController.GetWarLog)
		api.GET("/clans/:tag", clanController.GetClan)
		api.GET("/players/:tag", playerController.GetPlayer)
		api.GET("/players/:tag/battle-log", playerController.GetBattleLog)

		rankingService := service.NewRankingService(cocapiClient, infraredis.NewRankingCache(runtimeConfig.Redis, 10*time.Minute))
		rankingController := controller.NewRankingController(rankingService)
		leagueService := service.NewLeagueService(cocapiClient, infraredis.NewLeagueCache(runtimeConfig.Redis, 30*time.Minute))
		leagueController := controller.NewLeagueController(leagueService)
		labelService := service.NewLabelService(cocapiClient, infraredis.NewLabelCache(runtimeConfig.Redis, time.Hour))
		labelController := controller.NewLabelController(labelService)

		api.GET("/clans/labels", labelController.GetClanLabels)
		api.GET("/players/labels", labelController.GetPlayerLabels)
		api.GET("/locations", rankingController.GetLocations)
		api.GET("/locations/:id/rankings/clans", rankingController.GetClanRanking)
		api.GET("/locations/:id/rankings/players", rankingController.GetPlayerRanking)
		api.GET("/locations/:id/rankings/clans-capital", rankingController.GetClanCapitalRanking)
		api.GET("/locations/:id/rankings/clans-builder-base", rankingController.GetClanBuilderBaseRanking)
		api.GET("/locations/:id/rankings/players-builder-base", rankingController.GetPlayerBuilderBaseRanking)
		api.GET("/leagues", leagueController.GetLeagues)
		api.GET("/leagues/:id", leagueController.GetLeague)
		api.GET("/leagues/:id/seasons", leagueController.GetLeagueSeasons)
		api.GET("/leagues/:id/seasons/:season/rankings", leagueController.GetLeagueSeasonRankings)
		api.GET("/leagues/:id/seasons/:season/tiers", leagueController.GetLeagueTiers)
		api.GET("/leagues/:id/seasons/:season/tiers/:tier", leagueController.GetLeagueTier)
		api.GET("/leagues/:id/seasons/:season/tiers/:tier/history", leagueController.GetLeagueHistory)
		api.GET("/war-leagues", leagueController.GetWarLeagues)
		api.GET("/war-leagues/:id", leagueController.GetWarLeague)

		capitalService := service.NewCapitalService(cocapiClient, infraredis.NewCapitalCache(runtimeConfig.Redis, 5*time.Minute, 30*time.Minute))
		capitalController := controller.NewCapitalController(capitalService)
		utilityService := service.NewUtilityService(cocapiClient, infraredis.NewUtilityCache(runtimeConfig.Redis, time.Hour))
		utilityController := controller.NewUtilityController(utilityService)

		api.GET("/clans/:tag/capital-raid-seasons", capitalController.GetCapitalRaidSeasons)
		api.GET("/capital-leagues", capitalController.GetCapitalLeagues)
		api.GET("/capital-leagues/:id", capitalController.GetCapitalLeague)
		api.GET("/builder-base-leagues", capitalController.GetBuilderBaseLeagues)
		api.GET("/builder-base-leagues/:id", capitalController.GetBuilderBaseLeague)
		api.GET("/gold-pass/current", utilityController.GetCurrentGoldPass)
		api.GET("/clans", utilityController.SearchClans)
		api.GET("/locations/:id", utilityController.GetLocation)
		api.POST("/players/:tag/verify-token", utilityController.VerifyPlayerToken)
		api.GET("/players/:tag/league-group", utilityController.GetPlayerLeagueGroup)

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
