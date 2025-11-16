package domain

type ReviewerStat struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Assignments int64  `json:"assignments"`
}

type StatsResponse struct {
	TotalPR   int64          `json:"total_pr"`
	OpenPR    int64          `json:"open_pr"`
	MergedPR  int64          `json:"merged_pr"`
	Reviewers []ReviewerStat `json:"reviewers"`
}
