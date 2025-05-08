package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/structxz/calc_v3/configs"
	"github.com/structxz/calc_v3/internal/constants"
	"github.com/structxz/calc_v3/internal/db/sqlite"
	"github.com/structxz/calc_v3/internal/logger"
	"github.com/structxz/calc_v3/internal/app"

	"go.uber.org/zap"
)

func main() {
	// Логгер
	opts := logger.DefaultOptions()
	opts.LogDir = "logs/orchestrator"

	log, err := logger.New(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrFailedInitLogger, err)
		os.Exit(1)
	}
	defer func() {
		if err := log.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, constants.ErrFailedSyncLogger, err)
		}
	}()

	// Конфигурация
	cfg, err := configs.NewServerConfig()
	if err != nil {
		log.Fatal(constants.ErrFailedInitConfig, zap.Error(err))
	}

	// Подключение к БД
	db, err := sqlite.New(log)
	if err != nil {
		log.Fatal(constants.ErrFailedOpenDB, zap.Error(err))
	}
	defer db.Close()

	if err := sqlite.RunMigrations(log, db.Db); err != nil {
		log.Fatal("Migration failed", zap.Error(err))
	}

	// Сервер
	srv := server.New(cfg, log, db)

	// Контекст завершения
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Запуск серверов
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal(constants.ErrFailedStartServer, zap.Error(err))
		}
	}()

	log.Info(constants.LogOrchestratorStarted)

	// Ожидание завершения
	<-ctx.Done()

	log.Info("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error(constants.ErrServerShutdownFailed, zap.Error(err))
	}

	log.Info(constants.LogOrchestratorStoppedGrace)
}
