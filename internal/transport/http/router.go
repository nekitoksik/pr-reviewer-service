package http

import (
	"log/slog"
	"pr-reviewer-service/internal/service"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	TeamService  *service.TeamService
	UserService  *service.UserService
	PRService    *service.PRService
	StatsService *service.StatsService
	Logger       *slog.Logger
}

func NewRouter(deps Dependencies) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	teamHandler := NewTeamHandler(deps.TeamService, deps.Logger)
	userHandler := NewUserHandler(deps.UserService, deps.Logger)
	prHandler := NewPullRequestHandler(deps.PRService, deps.Logger)
	statsHandler := NewStatsHandler(deps.StatsService, deps.Logger)

	r.GET("/health", func(c *gin.Context) {
		c.Status(200)
	})

	// Teams
	r.POST("/team/add", teamHandler.AddTeam)
	r.GET("/team/get", teamHandler.GetTeam)

	// Users
	r.POST("/users/setIsActive", userHandler.SetIsActive)
	r.GET("/users/getReview", userHandler.GetReview)

	// PullRequests
	r.POST("/pullRequest/create", prHandler.Create)
	r.POST("/pullRequest/merge", prHandler.Merge)
	r.POST("/pullRequest/reassign", prHandler.Reassign)

	r.GET("/stats", statsHandler.GetStats)
	// swagger
	registerSwagger(r)

	return r
}
