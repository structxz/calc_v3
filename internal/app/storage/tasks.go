// Package storage provides data storage functionalities for expressions and tasks.
package storage

import (
	"fmt"
	"time"

	"distributed_calculator/internal/constants"
	"distributed_calculator/internal/app/models"

	"go.uber.org/zap"
)

// SaveTask saves a task to storage and adds it to the task queue.
func (s *Storage) SaveTask(task *models.Task) error {
	if task.ID == "" {
		s.logger.Error("Failed to save task: empty ID")
		return fmt.Errorf("task ID cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	task.CreatedAt = now

	taskCopy := *task
	s.tasks.Store(task.ID, &taskCopy)
	s.taskQueue = append(s.taskQueue, taskCopy)

	s.logger.Info("Task saved successfully",
		zap.String("id", task.ID),
		zap.String(constants.FieldExpressionID, task.ExpressionID),
		zap.String(constants.FieldOperation, task.Operation))
	return nil
}

// GetTask retrieves a task by ID.
func (s *Storage) GetTask(id string) (*models.Task, error) {
	if value, ok := s.tasks.Load(id); ok {
		s.logger.Debug(constants.LogTaskRetrieved,
			zap.String("id", id))
		return value.(*models.Task), nil
	}
	s.logger.Warn("Task not found",
		zap.String("id", id))
	return nil, fmt.Errorf("task not found") // Исправлено на константную строку вместо strings.ToLower
}

// UpdateTaskResult updates a task's result and checks for expression completion.
func (s *Storage) UpdateTaskResult(id string, result float64) error {
	if value, ok := s.tasks.Load(id); ok {
		task := value.(*models.Task)
		task.Result = &result
		s.tasks.Store(id, task)
		s.logger.Info("Task result updated",
			zap.String("id", id),
			zap.Float64("result", result))

		allTasksCompleted := true
		s.tasks.Range(func(_, v interface{}) bool {
			t := v.(*models.Task)
			if t.ExpressionID == task.ExpressionID && t.Result == nil {
				allTasksCompleted = false
				return false
			}
			return true
		})

		if allTasksCompleted {
			if err := s.UpdateExpressionStatus(task.ExpressionID, models.StatusComplete); err != nil {
				s.logger.Error("Failed to update expression status",
					zap.String("expressionID", task.ExpressionID),
					zap.Error(err))
			}
		}

		return nil
	}
	s.logger.Error("Failed to update task result: task not found",
		zap.String("id", id))
	return fmt.Errorf("task not found") // Исправлено на константную строку вместо strings.ToLower
}

// GetNextTask retrieves and removes the next task from the queue.
func (s *Storage) GetNextTask() (*models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.taskQueue) == 0 {
		s.logger.Debug("No tasks available in queue")
		return nil, fmt.Errorf("task not found")
	}

	task := s.taskQueue[0]
	s.taskQueue = s.taskQueue[1:]

	s.logger.Info("Next task retrieved from queue",
		zap.String("id", task.ID),
		zap.String(constants.FieldExpressionID, task.ExpressionID))
	return &task, nil
}
