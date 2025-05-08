package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/structxz/calc_v3/configs"
	"github.com/structxz/calc_v3/internal/constants"
	"github.com/structxz/calc_v3/internal/logger"
	pb "github.com/structxz/calc_v3/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.uber.org/zap"
)

type Agent struct {
	config     *configs.WorkerConfig
	logger     *logger.Logger
	grpcClient pb.OrchestratorClient
	conn       *grpc.ClientConn
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	ID         string
}

// New создает нового агента.
func New(cfg *configs.WorkerConfig, log *logger.Logger) *Agent {
	ctx, cancel := context.WithCancel(context.Background())
	return &Agent{
		config: cfg,
		logger: log,
		ctx:    ctx,
		cancel: cancel,
		ID: fmt.Sprintf("%d", time.Now().UnixNano()),
	}
}

// Start запускает агента
func (a *Agent) Start() error {
	a.logger.Info("Connecting to orchestrator via gRPC")
	
	conn, err := grpc.Dial(
		a.config.OrchestratorURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		a.logger.Error("Failed to connect to orchestrator",
		zap.Error(err))
		return err
	}
	
	a.conn = conn
	a.grpcClient = pb.NewOrchestratorClient(conn)
	
	a.logger.Info("Starting agent",
		zap.Int(constants.FieldComputingPower, a.config.ComputingPower),
		zap.String(constants.FieldOrchestratorURL, a.config.OrchestratorURL))

	for i := 0; i < a.config.ComputingPower; i++ {
		a.wg.Add(1)
		go a.worker(i)
	}

	return nil
}

// Stop останавливает.
func (a *Agent) Stop() {
	a.cancel()
	a.wg.Wait()
	if a.conn != nil {
		a.conn.Close()
	}
	a.logger.Info("Agent stopped")
}
