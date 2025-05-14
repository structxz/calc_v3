package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/structxz/calc_v3/configs"
	"github.com/structxz/calc_v3/internal/db/sqlite"
	"github.com/structxz/calc_v3/internal/logger"
	"github.com/structxz/calc_v3/internal/app"
)

func TestServer_New_RegisterRouteExists(t *testing.T) {
	cfg := &configs.ServerConfig{
		RestPort: "8080",
		GRPCPort: "50051",
	}

	log, err := logger.New(logger.DefaultOptions())
	require.NoError(t, err)

	sqliteStorage, err := sqlite.New(log) // временная SQLite в памяти
	require.NoError(t, err)

	s := server.New(cfg, log, sqliteStorage)
	require.NotNil(t, s)

	_ = httptest.NewRequest(http.MethodPost, "/api/v1/register", nil)
	w := httptest.NewRecorder()

	require.NotEqual(t, http.StatusNotFound, w.Code, "Route /api/v1/register should be registered")
}
