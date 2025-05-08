package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/structxz/calc_v3/pkg/api"
	"github.com/structxz/calc_v3/configs"
	"github.com/structxz/calc_v3/internal/orchestrator"
	"github.com/structxz/calc_v3/internal/db/sqlite"
	"github.com/structxz/calc_v3/internal/logger"
	"github.com/structxz/calc_v3/internal/middleware"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	config   *configs.ServerConfig
	sqlite   *sqlite.SQLiteStorage
	logger   *logger.Logger
	restSrv  *http.Server
	grpcSrv  *grpc.Server
}

// New создаёт REST + gRPC сервер
func New(cfg *configs.ServerConfig, log *logger.Logger, sqliteStorage *sqlite.SQLiteStorage) *Server {
	s := &Server{
		config: cfg,
		logger: log,
		sqlite: sqliteStorage,
	}

	// --- REST setup ---
	router := mux.NewRouter()

	// Public
	public := router.PathPrefix("/api/v1").Subrouter()
	public.HandleFunc("/register", s.handleRegister).Methods(http.MethodPost)
	public.HandleFunc("/login", s.handleLogin).Methods(http.MethodPost)

	// Protected
	protected := router.PathPrefix("/api/v1").Subrouter()
	protected.Use(middleware.AuthMiddleware(s.logger))
	protected.HandleFunc("/calculate", s.handleCalculate).Methods(http.MethodPost)
	protected.HandleFunc("/expressions", s.handleListExpressions).Methods(http.MethodGet)
	protected.HandleFunc("/expressions/{id}", s.handleGetExpression).Methods(http.MethodGet)

	s.restSrv = &http.Server{
		Addr:         ":" + cfg.RestPort,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return s
}

// Start запускает gRPC и REST серверы параллельно
func (s *Server) Start() error {
	// Запускаем gRPC в отдельной горутине
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.config.GRPCPort))
		if err != nil {
			s.logger.Fatal("Failed to listen on gRPC", zap.Error(err))
		}

		s.grpcSrv = grpc.NewServer()
		api.RegisterOrchestratorServer(s.grpcSrv, orchestrator.New(s.logger, s.sqlite))

		s.logger.Info("gRPC server started", zap.String("port", s.config.GRPCPort))
		if err := s.grpcSrv.Serve(lis); err != nil {
			s.logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Запускаем HTTP сервер (блокирует)
	s.logger.Info("REST server started", zap.String("port", s.config.RestPort))
	return s.restSrv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.restSrv.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shut down REST server", zap.Error(err))
	}

	if s.grpcSrv != nil {
		s.grpcSrv.GracefulStop()
		s.logger.Info("gRPC server stopped gracefully")
	}

	return nil
}
