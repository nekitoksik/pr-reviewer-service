package dto

import (
	"errors"
	"pr-reviewer-service/internal/domain"
)

// dto for request /pullRequest/create
type CreatePullRequestRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

// dto for response /pullRequest/create and /pullRequest/merge
type PullRequestResponse struct {
	PR domain.PullRequest `json:"pr"`
}

// dto for request /pullRequest/merge
type MergePullRequestRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

// dto for request /pullRequest/reassign
type ReassignReviewerRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

// dto for response /pullRequest/reassign
type ReassignReviewerResponse struct {
	PR         domain.PullRequest `json:"pr"`
	ReplacedBy string             `json:"replaced_by"`
}

func (r *CreatePullRequestRequest) Validate() error {
	if r.PullRequestID == "" {
		return errors.New("pull_request_id is required")
	}
	if r.PullRequestName == "" {
		return errors.New("pull_request_name is required")
	}
	if r.AuthorID == "" {
		return errors.New("author_id is required")
	}
	return nil
}

func (r *MergePullRequestRequest) Validate() error {
	if r.PullRequestID == "" {
		return errors.New("pull_request_id is required")
	}
	return nil
}

func (r *ReassignReviewerRequest) Validate() error {
	if r.PullRequestID == "" {
		return errors.New("pull_request_id is required")
	}
	if r.OldUserID == "" {
		return errors.New("old_user_id is required")
	}
	return nil
}
