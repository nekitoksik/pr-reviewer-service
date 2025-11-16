package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"pr-reviewer-service/config"
	"pr-reviewer-service/internal/db"
	"pr-reviewer-service/internal/logger"
	"pr-reviewer-service/internal/repo"
	"pr-reviewer-service/internal/service"
	httptransport "pr-reviewer-service/internal/transport/http"
)

func Run() {
	//config
	cfg := config.MustLoad()

	//logger
	log := logger.New(&cfg.Server)

	log.Info("starting application",
		slog.String("server_port", strconv.Itoa(cfg.Server.Port)),
	)

	//Migrations
	log.Info("Starting Migrations...")
	if err := db.RunMigrations(cfg); err != nil {
		log.Error("Failed to make migrations", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("Migrations successfully completed")

	//Postgres connect
	ctx := context.Background()
	pool, err := db.New(ctx, cfg)
	if err != nil {
		log.Error("Failed connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	//repos
	log.Info("Initializing repositories...")
	teamRepo := repo.NewTeamRepo(pool)
	userRepo := repo.NewUserRepo(pool)
	prRepo := repo.NewPullRequestRepo(pool)
	statsRepo := repo.NewStatsRepo(pool)
	log.Info("Successfully initialized repositories")

	//services
	log.Info("Initializing services...")
	teamService := service.NewTeamService(*teamRepo, *userRepo)
	userService := service.NewUserService(*userRepo, *prRepo)
	prService := service.NewPRService(*prRepo, *userRepo, *teamRepo)
	statsService := service.NewStatsService(*statsRepo)

	r := httptransport.NewRouter(httptransport.Dependencies{
		TeamService:  teamService,
		UserService:  userService,
		PRService:    prService,
		StatsService: statsService,
		Logger:       log,
	})

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Server.Port),
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("http server listening", slog.String("addr", srv.Addr))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server listen failed", slog.Any("error", err))
		}
	}()

	<-quit
	log.Info("shutting down server...")

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Error("server forced to shutdown", slog.Any("error", err))
	} else {
		log.Info("server exited gracefully")
	}

}
