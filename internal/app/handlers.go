package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/structxz/calc_v3/internal/app/models"
	"github.com/structxz/calc_v3/internal/auth"
	"github.com/structxz/calc_v3/internal/constants"
	"github.com/structxz/calc_v3/internal/db/sqlite"
	"github.com/structxz/calc_v3/internal/jwtutil"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

func (s *Server) handleCalculate(w http.ResponseWriter, r *http.Request) {
	var req models.CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Failed to decode request body",
			zap.Error(err))
		s.writeError(w, http.StatusUnprocessableEntity, constants.ErrInvalidRequestBody)
		return
	}

	if req.Expression == "" {
		s.logger.Warn("Empty expression received")
		s.writeError(w, http.StatusUnprocessableEntity, constants.ErrInvalidRequestBody)
		return
	}

	_, err := s.parseExpression(req.Expression)
	if err != nil {
		s.logger.Error(constants.LogFailedParseExpression,
			zap.String(constants.FieldExpression, req.Expression),
			zap.Error(err))

		s.writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	expr := &models.Expression{
		ID:         uuid.New().String(),
		Expression: req.Expression,
		Status:     models.StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = s.sqlite.SaveExpression(s.logger, expr)
	if err != nil {
		s.logger.Error(constants.ErrFailedSaveExpression,
			zap.String(constants.FieldExpression, req.Expression),
			zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, constants.ErrFailedSaveExpression)
		return
	}

	s.logger.Info("Expression received for calculation",
		zap.String("id", expr.ID),
		zap.String(constants.FieldExpression, expr.Expression))

	go func() {
		if err := s.processExpression(expr); err != nil {
			s.logger.Error("Failed to process expression",
				zap.String("id", expr.ID),
				zap.String(constants.FieldExpression, expr.Expression),
				zap.Error(err))

			if updateErr := s.sqlite.UpdateExpressionError(s.logger, expr.ID, err.Error()); updateErr != nil {
				s.logger.Error("Failed to update expression error status",
					zap.String("id", expr.ID),
					zap.Error(updateErr))
			}
		}
	}()

	s.writeJSON(w, http.StatusCreated, models.CalculateResponse{ID: expr.ID})
}

func (s *Server) handleListExpressions(w http.ResponseWriter, _ *http.Request) {
	expressions, err := s.sqlite.ListExpressions(s.logger)
	if err != nil {
		s.logger.Error("Failed to list expressions", zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, "failed to fetch expressions")
		return
	}

	s.logger.Info("Listing all expressions",
		zap.Int(constants.FieldCount, len(expressions)))

	s.writeJSON(w, http.StatusOK, models.ExpressionsResponse{Expressions: expressions})
}

func (s *Server) handleGetExpression(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	expr, err := s.sqlite.GetExpression(s.logger, id)
	if err != nil {
		s.logger.Error(constants.ErrFailedGetExpression,
			zap.String("id", id),
			zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, constants.ErrFailedGetExpression)
		return
	}
	if expr == nil {
		s.logger.Warn(constants.ErrExpressionNotFound,
			zap.String(constants.FieldID, id))
		s.writeError(w, http.StatusNotFound, constants.ErrExpressionNotFound)
		return
	}

	s.logger.Info("Expression retrieved",
		zap.String("id", id),
		zap.String("status", string(expr.Status)),
		zap.String("result", fmt.Sprintf("%v", *expr.Result)))
	s.writeJSON(w, http.StatusOK, models.ExpressionResponse{Expression: *expr})
}

func (s *Server) handleGetTask(w http.ResponseWriter, _ *http.Request) {
	task, err := s.sqlite.GetNextTask(s.logger)
	if err != nil {
		s.logger.Debug(constants.LogNoTasksAvailable)
		s.writeError(w, http.StatusNotFound, constants.ErrTaskNotFound)
		return
	}

	if task == nil {
		s.logger.Debug(constants.LogNoTasksAvailable)
		s.writeError(w, http.StatusNotFound, "No task available")
		return
	}

	s.logger.Info(constants.LogTaskRetrieved,
		zap.String(constants.FieldTaskID, task.ID),
		zap.String(constants.FieldOperation, task.Operation))
	s.writeJSON(w, http.StatusOK, models.TaskResponse{Task: *task})
}

func (s *Server) handleSubmitTaskResult(w http.ResponseWriter, r *http.Request) {
	var req models.TaskResult
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error("Failed to decode task result request", zap.Error(err))
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ID == "" {
		s.writeError(w, http.StatusBadRequest, "missing task ID")
		return
	}

	if err := s.sqlite.UpdateTaskResult(s.logger, req.ID, req.Result); err != nil {
		s.logger.Error("Failed to update task result",
			zap.String("task_id", req.ID),
			zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, "failed to update task result")
		return
	}

	s.logger.Info("Task result submitted",
		zap.String("task_id", req.ID),
		zap.Float64("result", req.Result))

	exprID, err := s.sqlite.GetExpressionIDByTaskID(req.ID)
	if err != nil {
		s.logger.Error("Failed to get expression ID by task ID", zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, "failed to find expression")
		return
	}

	done, err := s.sqlite.AreAllTasksCompleted(s.logger, exprID)
	if err != nil {
		s.logger.Error("Failed to check if all tasks are complete", zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if done {
		if err := s.sqlite.UpdateExpressionStatus(s.logger, exprID, models.StatusComplete); err != nil {
			s.logger.Error("Failed to update expression status to done", zap.Error(err))
			s.writeError(w, http.StatusInternalServerError, "failed to finalize expression")
			return
		}

		result, err := s.sqlite.GetFinalTaskResult(exprID)
		if err != nil {
			s.logger.Error("Failed to get final result", zap.Error(err))
			return
		}

		if err := s.sqlite.UpdateExpressionResult(s.logger, exprID, result); err != nil {
			s.logger.Error("Failed to update expression result", zap.Error(err))
		}

		s.logger.Info("All tasks completed — expression marked as done",
			zap.String("expression_id", exprID))
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	request, err := checkRightCreds(w, r, s)
	if err != nil {
		return
	}

	hashedPassword, err := auth.HashPassword(request.Password)
	if err != nil {
		s.logger.Error(constants.ErrFailedHashPassword,
			zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, constants.ErrFailedHashPassword)
		return
	}

	user := &models.User{
		Login:    request.Login,
		Password: hashedPassword,
	}

	storage, err := sqlite.New(s.logger)
	if err != nil {
		s.logger.Error(constants.ErrFailedOpenDB,
			zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, constants.ErrFailedOpenDB)
		return
	}
	defer storage.Close()

	err = storage.InsertUser(s.logger, user)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			s.logger.Warn(constants.ErrAlreadyExistUserInDB)
			s.writeError(w, http.StatusBadRequest, fmt.Sprintf("User with login %s already exists", user.Login))
			return
		}
		s.logger.Error(constants.ErrFailedInsertUser,
			zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, constants.ErrFailedInsertUser)
		return
	}

	s.logger.Info(constants.LogRegistered,
		zap.String(constants.FieldLogin, user.Login),
		zap.String(constants.FieldPassword, hashedPassword))

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	request, err := checkRightCreds(w, r, s)
	if err != nil {
		return
	}

	storage, err := sqlite.New(s.logger)
	if err != nil {
		s.logger.Error(constants.ErrFailedOpenDB,
			zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, constants.ErrFailedOpenDB)
		return
	}
	defer storage.Close()

	foundUser, err := storage.SelectUser(s.logger, request.Login)
	if err != nil {
		s.logger.Error(constants.ErrFailedSelectUser,
			zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, constants.ErrFailedSelectUser)
		return
	}

	if (foundUser == nil) || (foundUser.Login != request.Login) || (!auth.CheckPassword(foundUser.Password, request.Password)) {
		if foundUser == nil {
			s.logger.Error(constants.ErrInvalidLoginPassword,
				zap.Error(err))
			s.writeError(w, http.StatusUnauthorized, constants.ErrInvalidLoginPassword)
			return
		}

		s.logger.Error(constants.ErrInvalidLoginPassword,
			zap.Error(err))
		s.writeError(w, http.StatusUnauthorized, constants.ErrInvalidLoginPassword)
		return
	}

	jwt, err := jwtutil.MakeJWT(request.Login)
	if err != nil {
		s.logger.Error(constants.ErrInvalidLoginPassword,
			zap.Error(err))
		s.writeError(w, http.StatusUnauthorized, constants.ErrInvalidLoginPassword)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    jwt,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	s.logger.Info(constants.LogAuthenticated,
		zap.String(constants.FieldLogin, request.Login),
		zap.String(constants.FieldPassword, foundUser.Password),
		zap.String(constants.FieldJWT, jwt))
	
	s.writeJSON(w, http.StatusOK, map[string]string{"token": jwt})
}

func checkRightCreds(w http.ResponseWriter, r *http.Request, s *Server) (models.RegisterRequest, error) {
	var request models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		s.logger.Error("Failed to decode request body",
			zap.Error(err))
		s.writeError(w, http.StatusBadRequest, constants.ErrInvalidRequestBody)
		return request, errors.New("some error")
	}

	if request.Login == "" {
		s.logger.Warn("Empty login received")
		s.writeError(w, http.StatusBadRequest, constants.ErrInvalidRequestBody)
		return request, errors.New("some error")
	}
	if request.Password == "" {
		s.logger.Warn("Empty password received")
		s.writeError(w, http.StatusBadRequest, constants.ErrInvalidRequestBody)
		return request, errors.New("some error")
	}
	return request, nil
}
