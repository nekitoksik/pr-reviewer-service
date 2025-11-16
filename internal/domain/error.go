package domain

import "errors"

type ErrorCode string

const (
	ErrorTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrorPRExists    ErrorCode = "PR_EXISTS"
	ErrorPRMerged    ErrorCode = "PR_MERGED"
	ErrorNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrorNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrorNotFound    ErrorCode = "NOT_FOUND"
)

// чтобы удобно было сравнивать через errors.Is
var (
	ErrNotFound    = errors.New("not found")
	ErrTeamExists  = errors.New("team already exists")
	ErrPRExists    = errors.New("pull request already exists")
	ErrPRMerged    = errors.New("pull request already merged")
	ErrNoCandidate = errors.New("no candidate")
	ErrNotAssigned = errors.New("not assigned to this PR")
)

type Error struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

type ErrorResponse struct {
	Error Error `json:"error"`
}
