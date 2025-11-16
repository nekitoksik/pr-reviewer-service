package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/service"
	"pr-reviewer-service/internal/transport/http/dto"
)

type TeamHandler struct {
	svc    *service.TeamService
	logger *slog.Logger
}

func NewTeamHandler(svc *service.TeamService, logger *slog.Logger) *TeamHandler {
	return &TeamHandler{
		svc:    svc,
		logger: logger,
	}
}

// POST /team/add
func (h *TeamHandler) AddTeam(c *gin.Context) {
	var req dto.TeamAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: domain.Error{
				Code:    domain.ErrorNotFound,
				Message: "invalid request body",
			},
		})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: domain.Error{
				Code:    domain.ErrorNotFound,
				Message: "invalid request body",
			},
		})
		return
	}

	team := domain.Team{
		TeamName: req.TeamName,
		Members:  req.Members,
	}

	created, err := h.svc.CreateTeam(c.Request.Context(), team)
	if err != nil {
		if errors.Is(err, domain.ErrTeamExists) {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error: domain.Error{
					Code:    domain.ErrorTeamExists,
					Message: "team already exists",
				},
			})
			return
		}

		h.logger.Error("failed to create team", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error: domain.Error{
				Code:    domain.ErrorNotFound,
				Message: "internal error",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, dto.TeamAddResponse{
		Team: created,
	})
}

// GET /team/get?team_name=
func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: domain.Error{
				Code:    domain.ErrorNotFound,
				Message: "team_name is required",
			},
		})
		return
	}

	team, err := h.svc.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Error: domain.Error{
					Code:    domain.ErrorNotFound,
					Message: "resource not found",
				},
			})
			return
		}

		h.logger.Error("failed to get team", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error: domain.Error{
				Code:    domain.ErrorNotFound,
				Message: "internal error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, team)
}
