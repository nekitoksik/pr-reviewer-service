package http

import (
	"errors"
	"log/slog"
	"net/http"
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/service"
	"pr-reviewer-service/internal/transport/http/dto"

	"github.com/gin-gonic/gin"
)

type PullRequestHandler struct {
	svc    *service.PRService
	logger *slog.Logger
}

func NewPullRequestHandler(svc *service.PRService, logger *slog.Logger) *PullRequestHandler {
	return &PullRequestHandler{svc: svc, logger: logger}
}

// POST /pullRequest/create
func (h *PullRequestHandler) Create(c *gin.Context) {
	var req dto.CreatePullRequestRequest

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
				Message: err.Error(),
			},
		})
		return
	}

	pr, err := h.svc.Create(c.Request.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPRExists):
			c.JSON(http.StatusConflict, domain.ErrorResponse{
				Error: domain.Error{
					Code:    domain.ErrorPRExists,
					Message: "PR id already exists",
				},
			})
			return
		case errors.Is(err, domain.ErrNotFound):
			// автор или команда не найдены
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Error: domain.Error{
					Code:    domain.ErrorNotFound,
					Message: "resource not found",
				},
			})
			return
		default:
			h.logger.Error("failed to create pull request", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error: domain.Error{
					Code:    domain.ErrorNotFound, // можешь завести INTERNAL отдельно
					Message: "internal error",
				},
			})
			return
		}
	}

	c.JSON(http.StatusCreated, dto.PullRequestResponse{
		PR: pr,
	})
}

// POST /pullRequest/merge
func (h *PullRequestHandler) Merge(c *gin.Context) {
	var req dto.MergePullRequestRequest

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
				Message: err.Error(),
			},
		})
		return
	}

	pr, err := h.svc.Merge(c.Request.Context(), req.PullRequestID)
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

		h.logger.Error("failed to merge pull request", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error: domain.Error{
				Code:    domain.ErrorNotFound,
				Message: "internal error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.PullRequestResponse{
		PR: pr,
	})
}

// POST /pullRequest/reassign
func (h *PullRequestHandler) Reassign(c *gin.Context) {
	var req dto.ReassignReviewerRequest

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
				Message: err.Error(),
			},
		})
		return
	}

	pr, replacedBy, err := h.svc.ReassignReviewer(c.Request.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrPRMerged):
			c.JSON(http.StatusConflict, domain.ErrorResponse{
				Error: domain.Error{
					Code:    domain.ErrorPRMerged,
					Message: "cannot reassign on merged PR",
				},
			})
			return

		case errors.Is(err, domain.ErrNotAssigned):
			c.JSON(http.StatusConflict, domain.ErrorResponse{
				Error: domain.Error{
					Code:    domain.ErrorNotAssigned,
					Message: "reviewer is not assigned to this PR",
				},
			})
			return
		case errors.Is(err, domain.ErrNoCandidate):
			c.JSON(http.StatusConflict, domain.ErrorResponse{
				Error: domain.Error{
					Code:    domain.ErrorNoCandidate,
					Message: "no active replacement candidate in team",
				},
			})
			return
		case errors.Is(err, domain.ErrNotFound):
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Error: domain.Error{
					Code:    domain.ErrorNotFound,
					Message: "resource not found",
				},
			})
			return
		default:
			h.logger.Error("failed to reassign reviewer", slog.Any("error", err))
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
				Error: domain.Error{
					Code:    domain.ErrorNotFound,
					Message: "internal error",
				},
			})
			return
		}
	}

	c.JSON(http.StatusOK, dto.ReassignReviewerResponse{
		PR:         pr,
		ReplacedBy: replacedBy,
	})
}
