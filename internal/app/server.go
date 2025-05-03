package server

import (
	"context"
	"net/http"
	"time"

	"distributed_calculator/internal/constants"
	"distributed_calculator/configs"
	"distributed_calculator/internal/logger"
	"distributed_calculator/internal/app/storage"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Server представляет собой HTTP-сервер с его конфигурацией, хранилищем и регистратором.
type Server struct {
	config  *configs.ServerConfig
	storage *storage.Storage
	logger  *logger.Logger
	server  *http.Server
}

// New creates a new Server instance with the provided configuration and logger.
func New(cfg *configs.ServerConfig, log *logger.Logger) *Server {
	s := &Server{
		config:  cfg,
		storage: storage.New(log.Logger),
		logger:  log,
	}

	router := mux.NewRouter()

	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/calculate", s.handleCalculate).Methods(http.MethodPost)
	api.HandleFunc("/expressions", s.handleListExpressions).Methods(http.MethodGet)
	api.HandleFunc("/expressions/{id}", s.handleGetExpression).Methods(http.MethodGet)
	api.HandleFunc("/register", s.handleRegister).Methods(http.MethodPost)
	api.HandleFunc("/login", s.handleLogin).Methods(http.MethodPost)

	internal := router.PathPrefix("/internal").Subrouter()
	internal.HandleFunc(constants.PathTask, s.handleGetTask).Methods(http.MethodGet)
	internal.HandleFunc(constants.PathTask, s.handleSubmitTaskResult).Methods(http.MethodPost)


	fs := http.FileServer(http.Dir("./web"))
	router.PathPrefix("/web/").Handler(http.StripPrefix("/web/", fs))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/calculate", http.StatusMovedPermanently)
	})

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

// GetHandler returns the HTTP handler for the server.
func (s *Server) GetHandler() http.Handler {
	return s.server.Handler
}

// Start begins listening on the configured port and serves HTTP requests.
func (s *Server) Start() error {
	s.logger.Info("Starting server", zap.String(constants.FieldPort, s.config.Port))
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}