package service

import (
	"context"
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/repo"
)

type UserService struct {
	users repo.User
	prs   repo.PullRequest
}

func NewUserService(users repo.User, prs repo.PullRequest) *UserService {
	return &UserService{
		users: users,
		prs:   prs,
	}
}

func (s *UserService) SetActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	if err := s.users.SetActive(ctx, userID, isActive); err != nil {
		return domain.User{}, err
	}

	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return domain.User{}, err
	}

	return u, nil
}

func (s *UserService) GetReviewPullRequests(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	if _, err := s.users.GetByID(ctx, userID); err != nil {
		return nil, err
	}

	prs, err := s.prs.GetByReviewer(ctx, userID)
	if err != nil {
		return nil, err
	}
	return prs, nil
}
