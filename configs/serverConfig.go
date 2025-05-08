package configs

import (
	"fmt"
	"os"
	"strconv"
)

type ServerConfig struct {
	RestPort          string // Port на котором будет прослушиваться REST сервер.
	GRPCPort          string // Port на котором будет прослушиваться gRPC сервер.
	TimeAdditionMS    int64  // Время в миллисекундах для операций сложения.
	TimeSubtractionMS int64  // Время в миллисекундах для операций вычитания.
	TimeMultiplyMS    int64  // Время в миллисекундах для операций умножения.
	TimeDivisionMS    int64  // Время в миллисекундах для операций деления.
}

func NewServerConfig() (*ServerConfig, error) {
	timeAdd, err := getEnvInt64("TIME_ADDITION_MS", 100)
	if err != nil {
		return nil, fmt.Errorf("invalid TIME_ADDITION_MS: %w", err)
	}

	timeSub, err := getEnvInt64("TIME_SUBTRACTION_MS", 100)
	if err != nil {
		return nil, fmt.Errorf("invalid TIME_SUBTRACTION_MS: %w", err)
	}

	timeMul, err := getEnvInt64("TIME_MULTIPLICATIONS_MS", 100)
	if err != nil {
		return nil, fmt.Errorf("invalid TIME_MULTIPLICATIONS_MS: %w", err)
	}

	timeDiv, err := getEnvInt64("TIME_DIVISIONS_MS", 100)
	if err != nil {
		return nil, fmt.Errorf("invalid TIME_DIVISIONS_MS: %w", err)
	}

	restPort := getEnvString("REST_PORT", "8080")

	grpcPort := getEnvString("GRPC_PORT", "50051")

	return &ServerConfig{
		RestPort:          restPort,
		GRPCPort:          grpcPort,
		TimeAdditionMS:    timeAdd,
		TimeSubtractionMS: timeSub,
		TimeMultiplyMS:    timeMul,
		TimeDivisionMS:    timeDiv,
	}, nil
}

func getEnvString(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvInt64(key string, defaultValue int64) (int64, error) {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue, nil
	}
	return strconv.ParseInt(value, 10, 64)
}
