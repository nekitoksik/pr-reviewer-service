package http

import (
	"log/slog"
	"net/http"
	"pr-reviewer-service/internal/service"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	statsService *service.StatsService
	logger       *slog.Logger
}

func NewStatsHandler(statsService *service.StatsService, logger *slog.Logger) *StatsHandler {
	return &StatsHandler{
		statsService: statsService,
		logger:       logger,
	}
}

func (h *StatsHandler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	stats, err := h.statsService.GetStats(ctx)
	if err != nil {
		h.logger.Error("failed to get stats", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}
