package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"distributed_calculator/internal/app/models"
	"distributed_calculator/internal/constants"
	"distributed_calculator/internal/db/sqlite"
	"distributed_calculator/internal/jwt"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"github.com/mattn/go-sqlite3"
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

	if err := s.storage.SaveExpression(expr); err != nil {
		s.logger.Error("Failed to save expression",
			zap.String(constants.FieldExpression, req.Expression),
			zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, constants.ErrFailedProcessExpression)
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

			if updateErr := s.storage.UpdateExpressionError(expr.ID, err.Error()); updateErr != nil {
				s.logger.Error("Failed to update expression error status",
					zap.String("id", expr.ID),
					zap.Error(updateErr))
			}
		}
	}()

	s.writeJSON(w, http.StatusCreated, models.CalculateResponse{ID: expr.ID})
}

func (s *Server) handleListExpressions(w http.ResponseWriter, _ *http.Request) {
	exprPointers := s.storage.ListExpressions()
	expressions := make([]models.Expression, len(exprPointers))
	for i, expr := range exprPointers {
		expressions[i] = *expr
	}
	s.logger.Debug("Listing all expressions",
		zap.Int(constants.FieldCount, len(expressions)))
	s.writeJSON(w, http.StatusOK, models.ExpressionsResponse{Expressions: expressions})
}

func (s *Server) handleGetExpression(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	expr, err := s.storage.GetExpression(id)
	if err != nil {
		s.logger.Warn(constants.LogExpressionRetrieved,
			zap.String("id", id))
		s.writeError(w, http.StatusNotFound, constants.ErrExpressionNotFound)
		return
	}

	s.logger.Debug(constants.LogExpressionRetrieved,
		zap.String("id", id),
		zap.String(constants.FieldStatus, string(expr.Status)))
	s.writeJSON(w, http.StatusOK, models.ExpressionResponse{Expression: *expr})
}

func (s *Server) handleGetTask(w http.ResponseWriter, _ *http.Request) {
	task, err := s.storage.GetNextTask()
	if err != nil {
		s.logger.Debug(constants.LogNoTasksAvailable)
		s.writeError(w, http.StatusNotFound, constants.ErrTaskNotFound)
		return
	}

	s.logger.Debug(constants.LogTaskRetrieved,
		zap.String(constants.FieldTaskID, task.ID),
		zap.String(constants.FieldOperation, task.Operation))
	s.writeJSON(w, http.StatusOK, models.TaskResponse{Task: *task})
}

func (s *Server) handleSubmitTaskResult(w http.ResponseWriter, r *http.Request) {
	var result models.TaskResult
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		s.logger.Error(constants.LogFailedDecodeTask, zap.Error(err))
		s.writeError(w, http.StatusUnprocessableEntity, constants.ErrInvalidRequestBody)
		return
	}

	if err := s.storage.UpdateTaskResult(result.ID, result.Result); err != nil {
		s.logger.Error(constants.LogFailedUpdateTask, zap.String(constants.FieldTaskID, result.ID), zap.Error(err))
		s.writeError(w, http.StatusNotFound, constants.ErrTaskNotFound)
		return
	}

	task, err := s.storage.GetTask(result.ID)
	if err != nil {
		s.logger.Error(constants.LogFailedGetTaskResult, zap.String(constants.FieldTaskID, result.ID), zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, constants.ErrFailedProcessResult)
		return
	}

	dependentTasks := s.storage.GetTasksByDependency(result.ID)
	for _, depTask := range dependentTasks {
		depTaskCopy := *depTask
		allDepsMet := true
		for _, depID := range depTask.DependsOnTaskIDs {
			depResult, err := s.storage.GetTaskResult(depID)
			if err != nil {
				allDepsMet = false
				break
			}
			if depTask.Arg1 == 0 {
				depTaskCopy.Arg1 = depResult
			} else if depTask.Arg2 == 0 {
				depTaskCopy.Arg2 = depResult
			}
		}
		if allDepsMet && depTaskCopy.Arg1 != 0 && depTaskCopy.Arg2 != 0 {
			if err := s.storage.SaveTask(&depTaskCopy); err != nil {
				s.logger.Error("Failed to update dependent task",
					zap.String(constants.FieldTaskID, depTaskCopy.ID),
					zap.String(constants.FieldExpressionID, depTaskCopy.ExpressionID),
					zap.Error(err))

				// Optionally update the parent expression with an error status
				if updateErr := s.storage.UpdateExpressionError(task.ExpressionID,
					"Failed to update dependent task: "+err.Error()); updateErr != nil {
					s.logger.Error("Failed to update expression error status",
						zap.String(constants.FieldExpressionID, task.ExpressionID),
						zap.Error(updateErr))
				}
			}
		}
	}

	allTasks := s.storage.GetTasksByExpressionID(task.ExpressionID)
	allCompleted := true
	for _, t := range allTasks {
		if _, err := s.storage.GetTaskResult(t.ID); err != nil {
			allCompleted = false
			break
		}
	}
	if allCompleted {
		if err := s.storage.UpdateExpressionResult(task.ExpressionID, result.Result); err != nil {
			s.logger.Error(constants.LogFailedUpdateExpr, zap.String(constants.FieldExpressionID, task.ExpressionID), zap.Error(err))
		}
	}

	s.logger.Info(constants.LogTaskProcessed,
		zap.String(constants.FieldTaskID, task.ID),
		zap.String(constants.FieldExpressionID, task.ExpressionID),
		zap.Float64(constants.FieldResult, result.Result))

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	request, err := checkRightCreds(w, r, s)
	if err != nil {
		return
	}

	user := models.User{
		Login:    request.Login,
		Password: request.Password,
	}

	db, ctx, err := sqlite.Open(s.logger)
	if err != nil {
		s.logger.Error(constants.ErrFailedSetDBConnection,
			zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, constants.ErrFailedSetDBConnection)
		return
	}
	defer db.Close()

	err = sqlite.CreateTables(ctx, db, s.logger)
	if err != nil {
		s.logger.Error(constants.ErrFailedCreateTables,
			zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, constants.ErrFailedCreateTables)
		return
	}

	err = sqlite.InsertUser(ctx, db, s.logger, &user)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			s.logger.Warn(constants.ErrAlreadyExistUserInDB)
			s.writeError(w, http.StatusBadRequest, constants.ErrAlreadyExistsUserLogin)
			return
		}
		s.logger.Error(constants.ErrFailedInsertUser,
			zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, constants.ErrFailedInsertUser)
		return
	}

	// TODO: add info logging
	s.logger.Info(constants.LogRegistered,
		zap.String(constants.FieldLogin, user.Login),
		zap.String(constants.FieldPassword, user.Password))

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	request, err := checkRightCreds(w, r, s)
	if err != nil {
		return
	}

	user := models.User{
		Login:    request.Login,
		Password: request.Password,
	}

	db, ctx, err := sqlite.Open(s.logger)
	if err != nil {
		s.logger.Error(constants.ErrFailedSetDBConnection,
			zap.Error(err))
		s.writeError(w, http.StatusInternalServerError, constants.ErrFailedSetDBConnection)
		return
	}
	defer db.Close()

	foundUser, err := sqlite.SelectUser(ctx, db, user.Login)
	if err != nil {
		s.logger.Warn(constants.ErrNoUserFound,
			zap.Error(err))
		s.writeError(w, http.StatusUnauthorized, constants.ErrNoUserFound)
		return
	}

	if (foundUser.Login != user.Login) || (foundUser.Password != user.Password) {
		s.logger.Error(constants.ErrInvalidLoginPassword,
			zap.Error(err))
		s.writeError(w, http.StatusUnauthorized, constants.ErrInvalidLoginPassword)
		return
	}

	jwt, err := jwt.MakeJWT(s.logger, user.Login)
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
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	// TODO: add info logging
	s.logger.Info(constants.LogAuthenticated,
		zap.String(constants.FieldLogin, user.Login),
		zap.String(constants.FieldPassword, user.Password),
		zap.String(constants.FieldJWT, jwt))

	w.WriteHeader(http.StatusOK)
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
