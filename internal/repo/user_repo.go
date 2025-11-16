package repo

import (
	"context"
	"errors"
	"pr-reviewer-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User interface {
	//создает или обновляет пользователей команды при /team/add
	UpsertTeamMembers(ctx context.Context, teamName string, members []domain.TeamMember) error

	//получить user по id
	GetByID(ctx context.Context, userID string) (domain.User, error)

	//получить всех активных пользователей команды
	GetActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error)

	//обновить флаг is_active /users/setIsActive
	SetActive(ctx context.Context, userID string, isActive bool) error
}

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{
		pool: pool,
	}
}

func (r *UserRepo) UpsertTeamMembers(ctx context.Context, teamName string, members []domain.TeamMember) error {
	if len(members) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	for _, m := range members {
		batch.Queue(
			`INSERT INTO users (user_id, username, team_name, is_active)
             VALUES ($1, $2, $3, $4)
             ON CONFLICT (user_id)
             DO UPDATE SET
                 username = EXCLUDED.username,
                 team_name = EXCLUDED.team_name,
                 is_active = EXCLUDED.is_active`,
			m.UserID, m.Username, teamName, m.IsActive,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range members {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, userID string) (domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT user_id, username, team_name, is_active
         FROM users
         WHERE user_id = $1`,
		userID,
	).Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return u, nil
}

func (r *UserRepo) GetActiveByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT user_id, username, team_name, is_active
         FROM users
         WHERE team_name = $1 AND is_active = true`,
		teamName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.UserID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepo) SetActive(ctx context.Context, userID string, isActive bool) error {
	cmdTag, err := r.pool.Exec(ctx,
		`UPDATE users SET is_active = $2 WHERE user_id = $1`,
		userID, isActive,
	)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
