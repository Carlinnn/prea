package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prea-framework/prea/config"
	"github.com/prea-framework/prea/core"
	"github.com/prea-framework/prea/database"
	"github.com/prea-framework/prea/example/internal/preas"
	"github.com/prea-framework/prea/logger"
	"github.com/prea-framework/prea/security"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		logger.Error("prea: startup failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load("config", "")
	if err != nil {
		return err
	}
	// The keys are in config/application.yaml; missing ones stay "".
	var port, mode, logLevel, logFormat, dbURL, jwtSecret string
	var dbOpts database.Options
	if err := errors.Join(
		cfg.Value("server.port", &port),
		cfg.Value("server.mode", &mode),
		cfg.Value("logging.level", &logLevel),
		cfg.Value("logging.format", &logFormat),
		cfg.Value("database.url", &dbURL),
		cfg.Value("database", &dbOpts),
		cfg.Value("security.jwt.secret", &jwtSecret),
	); err != nil {
		return err
	}
	if err := logger.Configure(logLevel, logFormat); err != nil {
		return err
	}
	logger.Info("prea: config loaded", "profile", cfg.Profile())

	db, err := connect(dbURL)
	if err != nil {
		return err
	}
	// Runs after the server finished its in-flight requests (App.Run returns
	// only after a graceful shutdown).
	defer db.Close()

	// prea migrate runs the app with this set: run the command, don't serve.
	// It comes before the app's own checks, like the JWT secret.
	if args, ok := os.LookupEnv(database.MigrateEnv); ok {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		return database.Command(ctx, db, dbOpts.Migrations, strings.Fields(args), os.Stdout)
	}

	if err := security.SetSignKey(jwtSecret); err != nil {
		return err
	}
	gin.SetMode(mode)
	app := core.App{
		Modules:     preas.Modules,
		Security:    preas.Security,
		Server:      preas.Server,
		Middlewares: preas.Middlewares,
		Globals:     []any{db},
		Config:      cfg,
		Setup:       database.Setup(db, dbOpts),
	}
	return app.Run(":" + cmp.Or(port, "8000"))
}

// connect fails at startup when the database is unreachable; pgxpool.New is lazy.
func connect(dsn string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("database: %w", err)
	}
	return db, nil
}
