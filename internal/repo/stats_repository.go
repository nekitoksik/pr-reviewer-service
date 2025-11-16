package repo

import (
	"context"
	"fmt"
	"pr-reviewer-service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Stats interface {
	GetPRCounts(ctx context.Context) (total, open, merged int64, err error)
	GetReviewerStats(ctx context.Context) ([]domain.ReviewerStat, error)
}

type StatsRepo struct {
	pool *pgxpool.Pool
}

func NewStatsRepo(pool *pgxpool.Pool) *StatsRepo {
	return &StatsRepo{pool: pool}
}

func (r *StatsRepo) GetPRCounts(ctx context.Context) (total, open, merged int64, err error) {
	const totalQuery = `SELECT COUNT(*) FROM pull_requests;`

	if err = r.pool.QueryRow(ctx, totalQuery).Scan(&total); err != nil {
		return 0, 0, 0, fmt.Errorf("GetPRCounts total: %w", err)
	}

	const statusQuery = `
SELECT 
  COUNT(*) FILTER (WHERE status = 'OPEN')  AS open_pr,
  COUNT(*) FILTER (WHERE status = 'MERGED') AS merged_pr
FROM pull_requests;
`

	if err = r.pool.QueryRow(ctx, statusQuery).Scan(&open, &merged); err != nil {
		return 0, 0, 0, fmt.Errorf("GetPRCounts by status: %w", err)
	}

	return total, open, merged, nil
}

func (r *StatsRepo) GetReviewerStats(ctx context.Context) ([]domain.ReviewerStat, error) {
	const query = `
SELECT 
  u.user_id,
  u.username,
  COUNT(prr.pull_request_id) AS assignments
FROM users u
LEFT JOIN pr_reviewers prr ON prr.reviewer_id = u.user_id
GROUP BY u.user_id, u.username
ORDER BY assignments DESC;
`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetReviewerStats query: %w", err)
	}
	defer rows.Close()

	var res []domain.ReviewerStat

	for rows.Next() {
		var s domain.ReviewerStat
		if err := rows.Scan(&s.UserID, &s.Username, &s.Assignments); err != nil {
			return nil, fmt.Errorf("GetReviewerStats scan: %w", err)
		}
		res = append(res, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetReviewerStats rows: %w", err)
	}

	return res, nil
}
