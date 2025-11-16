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

type UserHandler struct {
	svc    *service.UserService
	logger *slog.Logger
}

func NewUserHandler(svc *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{svc: svc, logger: logger}
}

// POST /users/setIsActive
func (h *UserHandler) SetIsActive(c *gin.Context) {
	var req dto.SetUserActiveRequest

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

	user, err := h.svc.SetActive(c.Request.Context(), req.UserID, req.IsActive)
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

		h.logger.Error("failed to set user active", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error: domain.Error{
				Code:    domain.ErrorNotFound,
				Message: "internal error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.UserResponse{
		User: user,
	})
}

// GET /users/getReview?user_id=...
func (h *UserHandler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error: domain.Error{
				Code:    domain.ErrorNotFound,
				Message: "user_id is required",
			},
		})
		return
	}

	prs, err := h.svc.GetReviewPullRequests(c.Request.Context(), userID)
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

		h.logger.Error("failed to get user reviews", slog.Any("error", err))
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error: domain.Error{
				Code:    domain.ErrorNotFound,
				Message: "internal error",
			},
		})
		return
	}

	resp := dto.GetUserReviewResponse{
		UserID:       userID,
		PullRequests: prs,
	}

	c.JSON(http.StatusOK, resp)
}
