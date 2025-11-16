package repo

import (
	"context"
	"errors"
	"pr-reviewer-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Team interface {
	Create(ctx context.Context, team domain.Team) error

	GetByName(ctx context.Context, teamName string) (domain.Team, error)

	//в целом необязательный метод, создал для проверки команды, чтобы не тянуть еще и участников
	Exists(ctx context.Context, teamName string) (bool, error)
}

type TeamRepo struct {
	pool *pgxpool.Pool
}

func NewTeamRepo(pool *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{
		pool: pool,
	}
}

func (r *TeamRepo) Create(ctx context.Context, team domain.Team) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO teams (team_name) VALUES ($1)`,
		team.TeamName,
	)
	return err
}

func (r *TeamRepo) Exists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM teams WHERE team_name = $1)`,
		teamName,
	).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *TeamRepo) GetByName(ctx context.Context, teamName string) (domain.Team, error) {
	var name string
	err := r.pool.QueryRow(ctx,
		`SELECT team_name FROM teams WHERE team_name = $1`,
		teamName,
	).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Team{}, domain.ErrNotFound
		}

		return domain.Team{}, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1`,
		teamName,
	)

	if err != nil {
		return domain.Team{}, err
	}

	defer rows.Close()

	members := make([]domain.TeamMember, 0)
	for rows.Next() {
		var m domain.TeamMember
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return domain.Team{}, err
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return domain.Team{}, err
	}

	return domain.Team{
		TeamName: name,
		Members:  members,
	}, nil
}
