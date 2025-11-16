-- индекс для поиска участников команды по названию команды
CREATE INDEX idx_users_team_name ON users (team_name);

-- индекс для поиска pull_request по ревьюверу (users/getReview)
CREATE INDEX idx_pr_reviewers_reviewer_id ON pr_reviewers (reviewer_id);
