package orchestrator

import (
	"context"
	"errors"

	"github.com/structxz/calc_v3/internal/constants"
	"github.com/structxz/calc_v3/internal/db/sqlite"
	"github.com/structxz/calc_v3/internal/logger"
	"github.com/structxz/calc_v3/pkg/api"

	"go.uber.org/zap"
)

type OrchestratorServer struct {
	api.UnimplementedOrchestratorServer
	log     *logger.Logger
	storage *sqlite.SQLiteStorage
}

func New(log *logger.Logger, storage *sqlite.SQLiteStorage) *OrchestratorServer {
	return &OrchestratorServer{
		log:     log,
		storage: storage,
	}
}

// GetTask выдает следующую доступную задачу агенту
func (s *OrchestratorServer) GetTask(ctx context.Context, info *api.AgentInfo) (*api.TaskResponse, error) {
	task, err := s.storage.GetNextTask(s.log)
	if err != nil {
		return nil, err
	}

	if task == nil {
		return &api.TaskResponse{
			HasTask: false,
		}, nil
	}

	// Собираем gRPC-структуру задачи
	respTask := &api.Task{
		Id:           task.ID,
		ExpressionId: task.ExpressionID,
		Operation:    task.Operation,
		Operands:     []float64{task.Arg1, task.Arg2},
		DependsOn:    task.DependsOnTaskIDs,
	}

	// Обновляем статус задачи (например, RUNNING)
	err = s.storage.UpdateTaskStatus(s.log, task.ID, "RUNNING")
	if err != nil {
		return nil, err
	}

	return &api.TaskResponse{
		HasTask: true,
		Task:    respTask,
	}, nil
}

// SubmitTaskResult принимает результат от агента и обновляет состояние задачи
func (s *OrchestratorServer) SubmitTaskResult(ctx context.Context, res *api.TaskResult) (*api.SubmitResponse, error) {
	if res == nil {
		return nil, errors.New("empty result")
	}

	err := s.storage.UpdateTaskResult(s.log, res.TaskId, res.Result)
	if err != nil {
		return nil, err
	}

	allDone, err := s.storage.AreAllTasksCompleted(s.log, res.ExpressionId)
	if err != nil {
		return nil, err
	}

	if allDone {
		finalResult, err := s.storage.GetFinalTaskResult(res.ExpressionId)
		if err != nil {
			s.log.Error("Failed to get final task result", zap.Error(err))
		} else {
			s.storage.UpdateExpressionResult(s.log, res.ExpressionId, finalResult)
			s.log.Info(constants.LogFinalResultReady,
				zap.String(constants.FieldExpressionID, res.ExpressionId),
				zap.Float64("result", finalResult),
			)
		}
	}

	return &api.SubmitResponse{Success: true}, nil
}
