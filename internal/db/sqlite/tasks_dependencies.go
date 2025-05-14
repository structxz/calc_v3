package sqlite

import (
	"github.com/structxz/calc_v3/internal/logger"

	"go.uber.org/zap"
)

func (s *SQLiteStorage) SaveTaskDependencies(logger *logger.Logger, taskID string, dependencyID string) error {
	query := `INSERT OR IGNORE INTO task_dependencies (task_id, dependency_id) VALUES (?, ?)`
	_, err := s.Db.Exec(query, taskID, dependencyID)
	if err != nil {
		logger.Error("Failed to save task dependency", zap.String("taskID", taskID), zap.String("dependencyID", dependencyID), zap.Error(err))
	}
	return err
}


func (s *SQLiteStorage) GetTaskDependencies(logger *logger.Logger, taskID string) ([]string, error) {
    query := `SELECT dependency_id FROM task_dependencies WHERE task_id = ?`
    rows, err := s.Db.Query(query, taskID)
    if err != nil {
        logger.Error("Failed to get task dependencies", zap.String("taskID", taskID), zap.Error(err))
        return nil, err
    }
    defer rows.Close()

    var dependencies []string
    for rows.Next() {
        var dependencyID string
        if err := rows.Scan(&dependencyID); err != nil {
            logger.Error("Failed to scan task dependency", zap.String("taskID", taskID), zap.Error(err))
            return nil, err
        }
        dependencies = append(dependencies, dependencyID)
    }

    return dependencies, nil
}

func (s *SQLiteStorage) DeleteTaskDependency(logger *logger.Logger, taskID string, dependencyID string) error {
    query := `DELETE FROM task_dependencies WHERE task_id = ? AND dependency_id = ?`
    if _, err := s.Db.Exec(query, taskID, dependencyID); err != nil {
        logger.Error("Failed to delete task dependency", zap.String("taskID", taskID), zap.String("dependencyID", dependencyID), zap.Error(err))
        return err
    }
    return nil
}
