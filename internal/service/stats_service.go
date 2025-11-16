package service

import (
	"context"
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/repo"
)

type StatsService struct {
	stats repo.StatsRepo
}

func NewStatsService(stats repo.StatsRepo) *StatsService {
	return &StatsService{stats: stats}
}

func (s *StatsService) GetStats(ctx context.Context) (*domain.StatsResponse, error) {
	total, open, merged, err := s.stats.GetPRCounts(ctx)
	if err != nil {
		return nil, err
	}

	reviewers, err := s.stats.GetReviewerStats(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.StatsResponse{
		TotalPR:   total,
		OpenPR:    open,
		MergedPR:  merged,
		Reviewers: reviewers,
	}, nil
}
