package server

import (
	"context"
	"net/http"
	"time"

	"distributed_calculator/configs"
	"distributed_calculator/internal/constants"
	"distributed_calculator/internal/db/sqlite"
	"distributed_calculator/internal/logger"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Server is an HTTP server with its configuration, storage, and registrar
type Server struct {
	config  *configs.ServerConfig
	sqlite  *sqlite.SQLiteStorage
	logger  *logger.Logger
	server  *http.Server
}

// New creates a new Server instance with the provided configuration and logger.
func New(cfg *configs.ServerConfig, log *logger.Logger, sqliteStorage *sqlite.SQLiteStorage) *Server {
	s := &Server{
		config:  cfg,
		logger:  log,
		sqlite: sqliteStorage,
	}

	router := mux.NewRouter()

	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/calculate", s.handleCalculate).Methods(http.MethodPost)
	api.HandleFunc("/expressions", s.handleListExpressions).Methods(http.MethodGet)
	api.HandleFunc("/expressions/{id}", s.handleGetExpression).Methods(http.MethodGet)
	api.HandleFunc("/register", s.handleRegister).Methods(http.MethodPost)
	api.HandleFunc("/login", s.handleLogin).Methods(http.MethodPost)

	internal := router.PathPrefix("/internal").Subrouter()
	internal.HandleFunc("/task", s.handleGetTask).Methods(http.MethodGet)
	internal.HandleFunc("/task", s.handleSubmitTaskResult).Methods(http.MethodPost)

	s.server = &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	s.logger.Info("Server initialized",
		zap.String(constants.FieldPort, cfg.Port),
		zap.Int64("timeAdditionMS", cfg.TimeAdditionMS),
		zap.Int64("timeSubtractionMS", cfg.TimeSubtractionMS),
		zap.Int64("timeMultiplyMS", cfg.TimeMultiplyMS),
		zap.Int64("timeDivisionMS", cfg.TimeDivisionMS))

	return s
}

func (s *Server) GetHandler() http.Handler {
	return s.server.Handler
}

func (s *Server) InitDB(logger *logger.Logger) *sqlite.SQLiteStorage {
	sqliteStorage, err := sqlite.New(logger)
	if err != nil {
		logger.Fatal("Could not initialize database",
			zap.Error(err))
	}

	// Run migrations
	if err := sqlite.RunMigrations(logger, sqliteStorage.Db); err != nil {
		logger.Fatal("Migration failed",
			zap.Error(err))
	}
	return sqliteStorage
}

// Start begins listening on the configured port and serves HTTP requests.
func (s *Server) Start() error {
	s.logger.Info("Starting server", zap.String(constants.FieldPort, s.config.Port))
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}