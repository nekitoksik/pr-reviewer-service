package dto

import (
	"errors"
	"pr-reviewer-service/internal/domain"
)

// dto for request /users/setIsActive
type SetUserActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

// dto for response /users/setIsActive
type UserResponse struct {
	User domain.User `json:"user"`
}

// dto for response /users/getReview
type GetUserReviewResponse struct {
	UserID       string                    `json:"user_id"`
	PullRequests []domain.PullRequestShort `json:"pull_requests"`
}

func (r *SetUserActiveRequest) Validate() error {
	if r.UserID == "" {
		return errors.New("user_id is required")
	}
	return nil
}
