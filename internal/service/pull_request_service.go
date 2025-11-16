package service

import (
	"context"
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/repo"
	"time"
)

type PRService struct {
	prs   repo.PullRequest
	users repo.User
	teams repo.Team
}

func NewPRService(prs repo.PullRequest, users repo.User, teams repo.Team) *PRService {
	return &PRService{
		prs:   prs,
		users: users,
		teams: teams,
	}
}

func (s *PRService) Create(ctx context.Context, prID, prName, authorID string) (domain.PullRequest, error) {

	exists, err := s.prs.Exists(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, err
	}
	if exists {
		return domain.PullRequest{}, domain.ErrPRExists
	}

	author, err := s.users.GetByID(ctx, authorID)
	if err != nil {
		return domain.PullRequest{}, err
	}
	if author.TeamName == "" {
		return domain.PullRequest{}, domain.ErrNotFound
	}

	candidates, err := s.users.GetActiveByTeam(ctx, author.TeamName)
	if err != nil {
		return domain.PullRequest{}, err
	}

	reviewers := make([]string, 0, 2)
	for _, u := range candidates {
		if u.UserID == authorID {
			continue
		}
		reviewers = append(reviewers, u.UserID)
		if len(reviewers) == 2 {
			break
		}
	}

	now := time.Now().UTC()
	pr := domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            domain.PullRequestStatusOpen,
		AssignedReviewers: reviewers,
		CreatedAt:         &now,
		MergedAt:          nil,
	}

	if err := s.prs.Create(ctx, pr); err != nil {
		return domain.PullRequest{}, err
	}

	return pr, nil
}

func (s *PRService) Merge(ctx context.Context, prID string) (domain.PullRequest, error) {
	pr, err := s.prs.GetByID(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, err
	}

	if pr.Status == domain.PullRequestStatusMerged {
		return pr, nil
	}

	now := time.Now().UTC()
	pr.Status = domain.PullRequestStatusMerged
	pr.MergedAt = &now

	if err := s.prs.Update(ctx, pr); err != nil {
		return domain.PullRequest{}, err
	}

	return pr, nil
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (domain.PullRequest, string, error) {
	pr, err := s.prs.GetByID(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, "", err
	}

	if pr.Status == domain.PullRequestStatusMerged {
		return domain.PullRequest{}, "", domain.ErrPRMerged
	}

	assigned := false
	for _, r := range pr.AssignedReviewers {
		if r == oldReviewerID {
			assigned = true
			break
		}
	}

	if !assigned {
		return domain.PullRequest{}, "", domain.ErrNotAssigned
	}

	reviewer, err := s.users.GetByID(ctx, oldReviewerID)
	if err != nil {
		return domain.PullRequest{}, "", err
	}
	if reviewer.TeamName == "" {
		return domain.PullRequest{}, "", domain.ErrNotFound
	}

	candidates, err := s.users.GetActiveByTeam(ctx, reviewer.TeamName)
	if err != nil {
		return domain.PullRequest{}, "", err
	}

	assignedSet := make(map[string]struct{}, len(pr.AssignedReviewers))
	for _, id := range pr.AssignedReviewers {
		assignedSet[id] = struct{}{}
	}

	var newReviewerID string
	for _, u := range candidates {
		if u.UserID == oldReviewerID {
			continue
		}
		if _, already := assignedSet[u.UserID]; already {
			continue
		}
		newReviewerID = u.UserID
		break
	}

	if newReviewerID == "" {
		return domain.PullRequest{}, "", domain.ErrNoCandidate
	}

	for i, id := range pr.AssignedReviewers {
		if id == oldReviewerID {
			pr.AssignedReviewers[i] = newReviewerID
			break
		}
	}

	if err := s.prs.ReassignReviewer(ctx, prID, oldReviewerID, newReviewerID); err != nil {
		return domain.PullRequest{}, "", err
	}

	return pr, newReviewerID, nil
}
