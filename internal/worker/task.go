package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/structxz/calc_v3/pkg/api"

	"github.com/structxz/calc_v3/internal/app/models"
)

func (a *Agent) getTask() (*models.Task, error) {
	ctx, cancel := context.WithTimeout(a.ctx, 3*time.Second)
	defer cancel()

	resp, err := a.grpcClient.GetTask(ctx, &api.AgentInfo{
		AgentId: a.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("gRPC GetTask failed: %w", err)
	}

	if !resp.HasTask {
		return nil, nil
	}

	t := resp.Task

	if len(t.Operands) < 2 {
		return nil, fmt.Errorf("task has insufficient operands")
	}

	return &models.Task{
		ID:           t.Id,
		ExpressionID: t.ExpressionId,
		Operation:    t.Operation,
		Arg1:         t.Operands[0],
		Arg2: t.Operands[1],
		Result: nil,
		Status: "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		DependsOnTaskIDs: t.DependsOn,
	}, nil
}

func (a *Agent) sendResult(task *models.Task, result float64) error {
	ctx, cancel := context.WithTimeout(a.ctx, 3*time.Second)
	defer cancel()

	_, err := a.grpcClient.SubmitTaskResult(ctx, &api.TaskResult{
		TaskId: task.ID,
		ExpressionId: task.ExpressionID,
		Result: result,
	})

	if err != nil {
		return fmt.Errorf("gRPC SubmitTaskResult failed: %w", err)
	}

	return nil
}
