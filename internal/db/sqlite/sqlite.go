package sqlite

import (
	"database/sql"
	"distributed_calculator/internal/constants"
	"distributed_calculator/internal/logger"
	"go.uber.org/zap"
)

type SQLiteStorage struct {
	Db *sql.DB
}

func New(logger *logger.Logger) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", "sqlite.db")
	if err != nil {
		logger.Error(constants.ErrFailedOpenDB,
			zap.Error(err))
		return nil, err
	}

	if err := db.Ping(); err != nil {
		logger.Error(constants.ErrFailedVerifyDBConnection,
			zap.Error(err))
		return nil, err
	}

	return &SQLiteStorage{
		Db: db,
	}, nil
}

func RunMigrations(logger *logger.Logger, db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS expressions (
		id TEXT PRIMARY KEY,
		expression TEXT NOT NULL,
		status TEXT NOT NULL,
		result REAL,
		error TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT UNIQUE PRIMARY KEY,
		expression_id TEXT NOT NULL,
		operation TEXT NOT NULL,
		arg1 REAL NOT NULL,
		arg2 REAL,
		result REAL,
		status TEXT NOT NULL,
		error TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (expression_id) REFERENCES expressions(id)
	);

	CREATE TABLE IF NOT EXISTS task_dependencies (
		task_id TEXT NOT NULL,
    	depends_on_task_id TEXT NOT NULL,
    	PRIMARY KEY (task_id, depends_on_task_id),
    	FOREIGN KEY (task_id) REFERENCES tasks(id),
    	FOREIGN KEY (depends_on_task_id) REFERENCES tasks(id)
	);	

	CREATE TABLE IF NOT EXISTS users(
		login TEXT NOT NULL COLLATE NOCASE UNIQUE,
		password TEXT NOT NULL
	);
	`

	_, err := db.Exec(schema)
	if err != nil {
		logger.Error("failed to run migrations",
			zap.Error(err))
		return err
	}

	logger.Info("Database migration completed successfully")
	return nil
}

func (db *SQLiteStorage) Close() error {
	return db.Db.Close()
}

func nullFloat(f float64) sql.NullFloat64 {
	return sql.NullFloat64{
		Float64: f,
		Valid:   true,
	}
}