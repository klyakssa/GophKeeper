package app

import (
	"context"
	"errors"
	"fmt"
	"gophkeeper/internal/config"
	"gophkeeper/internal/logger"
	"gophkeeper/internal/repository/postgres"
	"gophkeeper/internal/service"
	httptransport "gophkeeper/internal/transport/http"
	client "gophkeeper/internal/transport/websocket"
	"gophkeeper/pkg/jwt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func Run(cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger := logger.NewLogger(cfg.Logging, cfg.App.Name)
	logger.Info("Config initialized:")
	logger.Info("Debug mode: " + fmt.Sprintf("%v", cfg.Debug))
	logger.Info("App name: " + cfg.App.Name)

	logger.Info("Initializing database connection...")
	repo, err := postgres.NewPostgresStorage(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize database connection", zap.Error(err))
		return
	}

	manager := client.NewManager(ctx, logger.Logger)

	logger.Info("Starting application...")

	// jwt
	if cfg.JWT.Secret == "" {
		logger.Error("JWT secret key is empty")
		return
	}
	jwtManager := jwt.NewJWTManager(cfg.JWT.Secret, cfg.JWT.Expire)

	// сервис
	authService := service.NewAuthService(repo, jwtManager)
	secureService := service.NewSecureService(repo)

	// handler
	authHandler := httptransport.NewAuthHandler(logger, authService)
	wsHandler := httptransport.NewWebSocketHandler(logger, authService, manager)
	secureHandler := httptransport.NewSecureHandler(secureService, logger, manager)

	// router
	router := httptransport.NewRouter(logger, cfg, jwtManager)
	router.RegisterRoutes(authHandler, wsHandler, secureHandler)

	go func() {
		if err := router.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("HTTP server stopped unexpectedly", zap.Error(err))
			cancel()
		}
	}()

	logger.Info("Application started")

	<-ctx.Done()
	logger.Info("Shutting down application...")

	if err := repo.Close(); err != nil {
		logger.Error("Failed to close database connection", zap.Error(err))
	}

	if err := manager.Close(); err != nil {
		logger.Error("Failed to close websocket connection", zap.Error(err))
	}

	logger.Info("Application terminated gracefully")
}
