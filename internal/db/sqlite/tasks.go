package sqlite

import (
	"database/sql"
	"distributed_calculator/internal/app/models"
	"distributed_calculator/internal/constants"
	"distributed_calculator/internal/logger"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// SaveTask - сохраняет задачу в таблицу tasks
func (s *SQLiteStorage) SaveTask(logger *logger.Logger, task *models.Task) error {
	query := `INSERT INTO tasks (id, expression_id, operation, arg1, arg2, result, status, created_at, updated_at)
	          VALUES (?, ?, ?, ?, ?, NULL, ?, ?, ?)`
	_, err := s.Db.Exec(query,
		task.ID,
		task.ExpressionID,
		task.Operation,
		nullFloat(task.Arg1),
		nullFloat(task.Arg2),
		task.Status,
		task.CreatedAt,
		time.Now(),
	)
	if err != nil {
		logger.Error("Failed to save task", zap.Error(err))
	}
	return err
}


// UpdateTaskStatus - обновляет статус задачи
func (s *SQLiteStorage) UpdateTaskStatus(logger *logger.Logger, id, status string) error {
	query := `UPDATE tasks SET status = ?, updated_at = ? WHERE id = ?`
	_, err := s.Db.Exec(query, status, time.Now(), id)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to update task status (task_id: %s)", id),
			zap.Error(err))
		return err
	}
	return nil 
}


// GetNextTask извлекает следующую задачу с состоянием PENDING и всеми зависимостями в статусе COMPLETED.
func (s *SQLiteStorage) GetNextTask(logger *logger.Logger) (*models.Task, error) {
	// SQL-запрос для получения задачи, которая не имеет незавершённых зависимостей.
	query := `
		SELECT id, expression_id, operation, arg1, arg2, status, created_at, updated_at
		FROM tasks
		WHERE status = 'PENDING'
		AND id NOT IN (
			SELECT td.task_id
			FROM task_dependencies td
			JOIN tasks dep ON td.depends_on_task_id = dep.id
			WHERE dep.status != 'COMPLETED'
		)
		LIMIT 1;
	`

	var task models.Task
	err := s.Db.QueryRow(query).Scan(
		&task.ID,
		&task.ExpressionID,
		&task.Operation,
		&task.Arg1,
		&task.Arg2,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если задачи нет — возвращаем nil.
			return nil, nil
		}
		// Возвращаем ошибку, если есть проблемы с запросом.
		logger.Error(fmt.Sprintf("failed to get next task: %v", err))
		return nil, fmt.Errorf("failed to get next task: %w", err)
	}

	// Логируем, что мы нашли задачу для выполнения.
	logger.Info(constants.LogTaskRetrieved,
		zap.String(constants.FieldTaskID, task.ID),
		zap.String(constants.FieldOperation, task.Operation))

	return &task, nil
}

// func (s *SQLiteStorage) CompleteTask(logger *logger.Logger, taskID string, result float64) error {
// 	query := `
// 		UPDATE tasks
// 		SET result = ?, status = ?, updated_at = ?
// 		WHERE id = ?`
// 	_, err := s.Db.Exec(query, result, models.StatusComplete, time.Now(), taskID)
// 	if err != nil {
// 		logger.Error("Failed to update task", zap.Error(err))
// 		return err
// 	}
// 	return nil
// }

// UpdateTaskResult обновляет результат выполнения задачи и статус
func (s *SQLiteStorage) UpdateTaskResult(logger *logger.Logger, taskID string, result float64) error {
	query := `UPDATE tasks SET result = ?, status = 'done', updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := s.Db.Exec(query, result, taskID)
	if err != nil {
		logger.Error("Failed to update task result", zap.String("task_id", taskID), zap.Error(err))
	}
	return err
}

// AreAllTasksCompleted проверяет, выполнены ли все задачи выражения
func (s *SQLiteStorage) AreAllTasksCompleted(logger *logger.Logger, exprID string) (bool, error) {
	query := `SELECT COUNT(*) FROM tasks WHERE expression_id = ? AND status != 'done'`
	var count int
	err := s.Db.QueryRow(query, exprID).Scan(&count)
	if err != nil {
		logger.Error("Failed to check task completion", zap.Error(err))
		return false, err
	}
	return count == 0, nil
}

func (s *SQLiteStorage) GetFinalTaskResult(expressionID string) (float64, error) {
	row := s.Db.QueryRow(`
		SELECT result FROM tasks 
		WHERE expression_id = ? AND status = 'done' 
		ORDER BY created_at DESC LIMIT 1
	`, expressionID)

	var result float64
	err := row.Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}