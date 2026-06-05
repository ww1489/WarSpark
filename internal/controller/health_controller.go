package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

	"github.com/ww1489/WarSpark/internal/utils"
)

type HealthController struct {
	mysql *sqlx.DB
	redis *redis.Client
}

func NewHealthController(mysql *sqlx.DB, redis *redis.Client) *HealthController {
	return &HealthController{mysql: mysql, redis: redis}
}

// Check reports whether the HTTP process is alive.
//
// @Summary Liveness check
// @Tags system
// @Produce json
// @Success 200 {object} utils.Response
// @Router /health [get]
// @Router /api/v1/health [get]
func (h *HealthController) Check(c *gin.Context) {
	utils.OK(c, gin.H{"status": "ok"})
}

// Ready checks infrastructure dependencies required to serve traffic.
//
// @Summary Readiness check
// @Tags system
// @Produce json
// @Success 200 {object} utils.Response
// @Failure 503 {object} utils.Response
// @Router /api/v1/ready [get]
func (h *HealthController) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	status := gin.H{
		"mysql": "ok",
		"redis": "ok",
	}

	if err := h.mysql.PingContext(ctx); err != nil {
		status["mysql"] = err.Error()
	}
	if err := h.redis.Ping(ctx).Err(); err != nil {
		status["redis"] = err.Error()
	}

	if status["mysql"] != "ok" || status["redis"] != "ok" {
		utils.JSON(c, http.StatusServiceUnavailable, 10006, "service unavailable", status)
		return
	}
	utils.OK(c, status)
}
