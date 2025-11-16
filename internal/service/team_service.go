package service

import (
	"context"
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/repo"
)

type TeamService struct {
	teams repo.Team
	users repo.User
}

func NewTeamService(teams repo.Team, users repo.User) *TeamService {
	return &TeamService{teams: teams, users: users}
}

func (s *TeamService) CreateTeam(ctx context.Context, team domain.Team) (domain.Team, error) {
	exists, err := s.teams.Exists(ctx, team.TeamName)
	if err != nil {
		return domain.Team{}, err
	}
	if exists {
		return domain.Team{}, domain.ErrTeamExists
	}

	if err := s.teams.Create(ctx, team); err != nil {
		return domain.Team{}, err
	}

	if err := s.users.UpsertTeamMembers(ctx, team.TeamName, team.Members); err != nil {
		return domain.Team{}, err
	}

	return team, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (domain.Team, error) {
	return s.teams.GetByName(ctx, teamName)
}
