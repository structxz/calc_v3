package sqlite

import (
	"database/sql"
	"github.com/structxz/calc_v3/internal/app/models"
	"github.com/structxz/calc_v3/internal/constants"
	"github.com/structxz/calc_v3/internal/logger"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

func (s *SQLiteStorage) SaveExpression(logger *logger.Logger, expr *models.Expression) error {
	query := `
		INSERT INTO expressions (id, expression, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?);
	`
	_, err := s.Db.Exec(query, expr.ID, expr.Expression, expr.Status, expr.CreatedAt, expr.UpdatedAt)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to insert expression (exp_id: %s)", expr.ID),
			zap.Error(err))
		return err
	}
	return nil
}

func (s *SQLiteStorage) GetExpression(logger *logger.Logger, id string) (*models.Expression, error) {
	query := `
		SELECT id, expression, status, result, created_at, updated_at, error
		FROM expressions
		WHERE id = ?
	`

	var expr models.Expression
	var result sql.NullFloat64
	var errorText sql.NullString
	var createdAt, updatedAt string

	err := s.Db.QueryRow(query, id).Scan(
		&expr.ID,
		&expr.Expression,
		&expr.Status,
		&result,
		&createdAt,
		&updatedAt,
		&errorText,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error(fmt.Sprintf("Failed to get expression (exp_id: %s)", expr.ID),
			zap.Error(err))
		return nil, err
	}

	if result.Valid {
		expr.Result = &result.Float64
	}
	if errorText.Valid {
		expr.Error = errorText.String
	}

	expr.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	expr.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &expr, nil
}

func (s *SQLiteStorage) UpdateExpressionResult(logger *logger.Logger, expressionID string, result float64) error {
	_, err := s.Db.Exec(
		`UPDATE expressions SET result = ?, status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		result, models.StatusComplete, expressionID,
	)
	if err != nil {
		logger.Error("Failed to update expression result", zap.String("expression_id", expressionID), zap.Error(err))
	}
	return err
}

func (s *SQLiteStorage) UpdateExpressionStatus(logger *logger.Logger, id string, status string) error {
	query := `UPDATE expressions SET status = ?, updated_at = ? WHERE id = ?`
	_, err := s.Db.Exec(query, status, time.Now(), id)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to update expression status (exp_id: %s)", id),
			zap.Error(err))
		return err
	}
	return nil
}

func (s *SQLiteStorage) UpdateExpressionError(logger *logger.Logger, id string, errorMsg string) error {
	const query = `
		UPDATE expressions
		SET status = ?, error = ?, updated_at = ?
		WHERE id = ?
	`
	_, err := s.Db.Exec(query, "failed", errorMsg, time.Now().UTC(), id)
	if err != nil {
		logger.Error(fmt.Sprintf("sqlite: failed to update expression error (exp_id: %s)", id),
			zap.Error(err))
		return err
	}

	return nil
}

func (s *SQLiteStorage) ListExpressions(logger *logger.Logger) ([]models.Expression, error) {
	query := `
		SELECT id, expression, status, result, created_at, updated_at, error
		FROM expressions
		ORDER BY created_at DESC
	`

	rows, err := s.Db.Query(query)
	if err != nil {
		logger.Error(constants.ErrFailedGetExpressions,
			zap.Error(err))
		return nil, fmt.Errorf("query expressions: %w", err)
	}
	defer rows.Close()

	var expressions []models.Expression
	for rows.Next() {
		var expr models.Expression
		var result sql.NullFloat64
		var errorText sql.NullString
		var createdAt, updatedAt string

		if err := rows.Scan(
			&expr.ID,
			&expr.Expression,
			&expr.Status,
			&result,
			&createdAt,
			&updatedAt,
			&errorText,
		); err != nil {
			logger.Error(fmt.Sprintf("failed to scan expression row: %v", err), zap.Error(err))
			continue
		}

		if result.Valid {
			expr.Result = &result.Float64
		}
		if errorText.Valid {
			expr.Error = errorText.String
		}
		expr.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		expr.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		expressions = append(expressions, expr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return expressions, nil
}

func (s *SQLiteStorage) GetExpressionIDByTaskID(taskID string) (string, error) {
	var exprID string
	query := `SELECT expression_id FROM tasks WHERE id = ?`
	err := s.Db.QueryRow(query, taskID).Scan(&exprID)
	if err != nil {
		return "", fmt.Errorf("failed to get expression_id: %w", err)
	}
	return exprID, nil
}
