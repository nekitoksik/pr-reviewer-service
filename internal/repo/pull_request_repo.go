package repo

import (
	"context"
	"errors"
	"pr-reviewer-service/internal/domain"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PullRequest interface {
	Create(ctx context.Context, pr domain.PullRequest) error

	GetByID(ctx context.Context, prID string) (domain.PullRequest, error)

	Update(ctx context.Context, pr domain.PullRequest) error

	Exists(ctx context.Context, prID string) (bool, error)

	GetByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error)

	GetReviewers(ctx context.Context, prID string) ([]string, error)

	ReassignReviewer(ctx context.Context, prID string, oldUserID, newReviewerID string) error

	SetReviewers(ctx context.Context, prID string, reviewersIDs []string) error
}

type PullRequestRepo struct {
	pool *pgxpool.Pool
}

func NewPullRequestRepo(pool *pgxpool.Pool) *PullRequestRepo {
	return &PullRequestRepo{
		pool: pool,
	}
}

func (r *PullRequestRepo) Exists(ctx context.Context, prID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`,
		prID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *PullRequestRepo) Create(ctx context.Context, pr domain.PullRequest) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO pull_requests
            (pull_request_id, pull_request_name, author_id, status, created_at, merged_at)
         VALUES ($1, $2, $3, $4, $5, $6)`,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		string(pr.Status),
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		return err
	}

	for _, reviewerID := range pr.AssignedReviewers {
		_, err = tx.Exec(ctx,
			`INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
             VALUES ($1, $2)`,
			pr.PullRequestID, reviewerID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PullRequestRepo) GetByID(ctx context.Context, prID string) (domain.PullRequest, error) {
	var pr domain.PullRequest
	var status string
	var createdAt, mergedAt *time.Time

	err := r.pool.QueryRow(ctx,
		`SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
         FROM pull_requests
         WHERE pull_request_id = $1`,
		prID,
	).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &status, &createdAt, &mergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.PullRequest{}, domain.ErrNotFound
		}
		return domain.PullRequest{}, err
	}

	pr.Status = domain.PullRequestStatus(status)
	pr.CreatedAt = createdAt
	pr.MergedAt = mergedAt

	reviewers, err := r.GetReviewers(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, err
	}
	pr.AssignedReviewers = reviewers

	return pr, nil
}

func (r *PullRequestRepo) Update(ctx context.Context, pr domain.PullRequest) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`UPDATE pull_requests
         SET pull_request_name = $2,
             author_id         = $3,
             status            = $4,
             created_at        = $5,
             merged_at         = $6
         WHERE pull_request_id = $1`,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		string(pr.Status),
		pr.CreatedAt,
		pr.MergedAt,
	)
	if err != nil {
		return err
	}

	// переустанавливаем ревьюверов
	_, err = tx.Exec(ctx,
		`DELETE FROM pr_reviewers WHERE pull_request_id = $1`,
		pr.PullRequestID,
	)
	if err != nil {
		return err
	}

	for _, reviewerID := range pr.AssignedReviewers {
		_, err = tx.Exec(ctx,
			`INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
             VALUES ($1, $2)`,
			pr.PullRequestID, reviewerID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PullRequestRepo) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT reviewer_id
         FROM pr_reviewers
         WHERE pull_request_id = $1`,
		prID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviewers []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return reviewers, nil
}

func (r *PullRequestRepo) SetReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`DELETE FROM pr_reviewers WHERE pull_request_id = $1`,
		prID,
	)
	if err != nil {
		return err
	}

	for _, reviewerID := range reviewerIDs {
		_, err = tx.Exec(ctx,
			`INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
             VALUES ($1, $2)`,
			prID, reviewerID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PullRequestRepo) ReassignReviewer(ctx context.Context, prID, oldReviewerID, newReviewerID string) error {
	cmdTag, err := r.pool.Exec(ctx,
		`UPDATE pr_reviewers
         SET reviewer_id = $3
         WHERE pull_request_id = $1 AND reviewer_id = $2`,
		prID, oldReviewerID, newReviewerID,
	)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound // NOT_ASSIGNED можно обрабатывать в сервисе как отдельный кейс
	}
	return nil
}

func (r *PullRequestRepo) GetByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT pr.pull_request_id,
                pr.pull_request_name,
                pr.author_id,
                pr.status
         FROM pull_requests pr
         JOIN pr_reviewers r ON r.pull_request_id = pr.pull_request_id
         WHERE r.reviewer_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.PullRequestShort
	for rows.Next() {
		var item domain.PullRequestShort
		var status string
		if err := rows.Scan(&item.PullRequestID, &item.PullRequestName, &item.AuthorID, &status); err != nil {
			return nil, err
		}
		item.Status = domain.PullRequestStatus(status)
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
