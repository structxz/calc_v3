package sqlite

import (
	"context"
	"database/sql"
	"distributed_calculator/internal/app/models"
	"distributed_calculator/internal/constants"
	"distributed_calculator/internal/logger"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

type DB struct {
	*sql.DB
}

func Open(logger *logger.Logger) (*sql.DB, context.Context, error){
	ctx := context.TODO()
	
	db, err := sql.Open("sqlite3", "users.db")
	if err != nil {
		logger.Error(constants.ErrFailedOpenDB,
			zap.Error(err))
		return nil, nil, err
	}

	err = db.PingContext(ctx)
	if err != nil {
		logger.Error(constants.ErrFailedVerifyDBConnection,
			zap.Error(err))
	}
	return db, ctx, nil
}

func CreateTables(ctx context.Context, db *sql.DB, logger *logger.Logger) error {
	const (
		usersTable = `
	CREATE TABLE IF NOT EXISTS users(
		login TEXT NOT NULL COLLATE NOCASE UNIQUE,
		password TEXT NOT NULL
	);`

		expressionsTable = `
	CREATE TABLE IF NOT EXISTS expressions(
		id UUID PRIMARY KEY,
    	expression TEXT,
    	status TEXT NOT NULL,
    	result DOUBLE PRECISION,
    	error TEXT
	);`
	)

	if _, err := db.ExecContext(ctx, usersTable); err != nil {
		logger.Error("Failed to create users table",
			zap.Error(err))
		return err
	}

	if _, err := db.ExecContext(ctx, expressionsTable); err != nil {
		logger.Error("Failed to create expressions table",
			zap.Error(err))
		return err
	}

	return nil
}

func InsertUser(ctx context.Context, db *sql.DB, logger *logger.Logger, user *models.User) error {
	var q = `
	INSERT INTO users (login, password) values ($1, $2)
	`
	_, err := db.ExecContext(ctx, q, user.Login, user.Password)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return err
		}
		logger.Error(fmt.Sprintf("Failed to insert user (login: %s)", user.Login),
			zap.Error(err))
		return err
	}

	return nil
}

func InsertExpression(ctx context.Context, db *sql.DB, expression *models.Expression) (int64, error) {
	var q = `
	INSERT INTO expressions (id, expression, status, result, error) values ($1, $2, $3, $4, $5)
	`
	result, err := db.ExecContext(ctx, q, expression.ID, expression.Expression, expression.Status, expression.Result, expression.Error)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func SelectUser(ctx context.Context, db *sql.DB, login string) (*models.User, error) {
	const query = `SELECT login, password FROM users WHERE login = ?`

    row := db.QueryRowContext(ctx, query, login)

    var user models.User
    err := row.Scan(&user.Login, &user.Password)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil // пользователь не найден
        }
        return nil, err // другая ошибка
    }

    return &user, nil
}
