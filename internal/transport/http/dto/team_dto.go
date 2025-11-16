package dto

import (
	"errors"
	"fmt"
	"pr-reviewer-service/internal/domain"
)

// dto for request /team/add
type TeamAddRequest struct {
	TeamName string              `json:"team_name"`
	Members  []domain.TeamMember `json:"members"`
}

// dto for response /team/add and /team/get
type TeamAddResponse struct {
	Team domain.Team `json:"team"`
}

func (r *TeamAddRequest) Validate() error {
	if r.TeamName == "" {
		return errors.New("team_name is required")
	}
	if len(r.Members) == 0 {
		return errors.New("members must not be empty")
	}
	for i, m := range r.Members {
		if m.UserID == "" {
			return fmt.Errorf("members[%d].user_id is required", i)
		}
		if m.Username == "" {
			return fmt.Errorf("members[%d].username is required", i)
		}
	}
	return nil
}
