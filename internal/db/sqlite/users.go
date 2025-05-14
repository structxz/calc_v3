package sqlite

import (
	"database/sql"
	"github.com/structxz/calc_v3/internal/app/models"
	"github.com/structxz/calc_v3/internal/logger"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func (s *SQLiteStorage) SelectUser(logger *logger.Logger, login string) (*models.User, error) {
	const query = `SELECT login, password FROM users WHERE login = ?`

	row := s.Db.QueryRow(query, login)

	var user models.User
	err := row.Scan(&user.Login, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		logger.Error(fmt.Sprintf("Failed to select user (login: %s)", login),
			zap.Error(err))
		return nil, err
	}

	return &user, nil
}

func (s *SQLiteStorage) InsertUser(log *logger.Logger, user *models.User) error {
	const query = `INSERT INTO users (login, password) VALUES (?, ?)`

	_, err := s.Db.Exec(query, user.Login, user.Password)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok &&
			sqliteErr.Code == sqlite3.ErrConstraint &&
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return err
		}

		log.Error(fmt.Sprintf("Failed to insert user (login: %s)", user.Login),
			zap.Error(err))
		return err
	}

	return nil
}