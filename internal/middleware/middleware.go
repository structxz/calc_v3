package middleware

import (
	"context"
	"distributed_calculator/internal/jwtutil"
	"distributed_calculator/internal/logger"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)



func AuthMiddleware(logger *logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			login, err := jwtutil.ValidateJWT(r)
			if err != nil {
				logger.Warn("Unauthorized request", zap.Error(err))
				http.Error(w, "You are unauthorized, first go through authentication", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), "user", login)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
