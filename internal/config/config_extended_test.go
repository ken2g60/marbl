package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear all flags and environment variables
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Clear relevant environment variables
	envVars := []string{
		"GRPC_PORT", "GRPC_ADDR", "PROMETHEUS_PORT", "DB_CONN",
		"MAX_BACKLOG", "LOG_LEVEL", "LOG_FORMAT", "PROFILING_PORT",
		"MESSAGE_RATE", "ZEROMQ_PORT", "ZEROMQ_ADDR",
	}
	for _, env := range envVars {
		os.Unsetenv(env)
	}

	// Mock command line args
	os.Args = []string{"cmd"}

	cfg := LoadConfig()

	// Test default values
	assert.Equal(t, "50051", cfg.GRPCPort)
	assert.Equal(t, "localhost:50051", cfg.GRPCAddr)
	assert.Equal(t, "9090", cfg.PrometheusPort)
	assert.Equal(t, "postgres://user:password@localhost:5432/tasks?sslmode=disable", cfg.DBConnString)
	assert.Equal(t, 1000, cfg.MaxBacklog)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.Equal(t, "6060", cfg.ProfilingPort)
	assert.Equal(t, 10, cfg.MessageRate)
	assert.Equal(t, "5555", cfg.ZeroMQPort)
	assert.Equal(t, "localhost:5555", cfg.ZeroMQAddr)
}

func TestLoadConfig_EnvironmentVariables(t *testing.T) {
	// Clear all flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Set environment variables
	envVars := map[string]string{
		"GRPC_PORT":       "60000",
		"GRPC_ADDR":       "remote:60000",
		"PROMETHEUS_PORT": "9091",
		"DB_CONN":         "postgres://test:test@localhost:5432/testdb?sslmode=disable",
		"MAX_BACKLOG":     "2000",
		"LOG_LEVEL":       "debug",
		"LOG_FORMAT":      "console",
		"PROFILING_PORT":  "6061",
		"MESSAGE_RATE":    "50",
		"ZEROMQ_PORT":     "6666",
		"ZEROMQ_ADDR":     "remote:6666",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	// Mock command line args
	os.Args = []string{"cmd"}

	cfg := LoadConfig()

	// Test environment variable values
	assert.Equal(t, "60000", cfg.GRPCPort)
	assert.Equal(t, "remote:60000", cfg.GRPCAddr)
	assert.Equal(t, "9091", cfg.PrometheusPort)
	assert.Equal(t, "postgres://test:test@localhost:5432/testdb?sslmode=disable", cfg.DBConnString)
	assert.Equal(t, 2000, cfg.MaxBacklog)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "console", cfg.LogFormat)
	assert.Equal(t, "6061", cfg.ProfilingPort)
	assert.Equal(t, 50, cfg.MessageRate)
	assert.Equal(t, "6666", cfg.ZeroMQPort)
	assert.Equal(t, "remote:6666", cfg.ZeroMQAddr)
}

func TestLoadConfig_CommandLineFlags(t *testing.T) {
	// Clear all flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Mock command line args with flags
	os.Args = []string{
		"cmd",
		"-grpc-port=70000",
		"-grpc-addr=flag:70000",
		"-prometheus-port=9092",
		"-db-conn=postgres://flag:flag@localhost:5432/flagdb?sslmode=disable",
		"-max-backlog=3000",
		"-log-level=warn",
		"-log-format=json",
		"-profiling-port=6062",
		"-message-rate=100",
		"-zeromq-port=7777",
		"-zeromq-addr=flag:7777",
	}

	cfg := LoadConfig()

	// Test flag values (flags should override environment variables and defaults)
	assert.Equal(t, "70000", cfg.GRPCPort)
	assert.Equal(t, "flag:70000", cfg.GRPCAddr)
	assert.Equal(t, "9092", cfg.PrometheusPort)
	assert.Equal(t, "postgres://flag:flag@localhost:5432/flagdb?sslmode=disable", cfg.DBConnString)
	assert.Equal(t, 3000, cfg.MaxBacklog)
	assert.Equal(t, "warn", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.Equal(t, "6062", cfg.ProfilingPort)
	assert.Equal(t, 100, cfg.MessageRate)
	assert.Equal(t, "7777", cfg.ZeroMQPort)
	assert.Equal(t, "flag:7777", cfg.ZeroMQAddr)
}

func TestLoadConfig_VersionFlag(t *testing.T) {
	// Clear all flags
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Mock command line args with version flag
	os.Args = []string{"cmd", "-version"}

	cfg := LoadConfig()

	// Test version flag
	assert.Equal(t, "version", cfg.Mode)
}

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		Mode:           "producer",
		GRPCPort:       "50051",
		GRPCAddr:       "localhost:50051",
		PrometheusPort: "9090",
		DBConnString:   "postgres://user:password@localhost:5432/tasks?sslmode=disable",
		MaxBacklog:     1000,
		LogLevel:       "info",
		LogFormat:      "json",
		ProfilingPort:  "6060",
		MessageRate:    10,
		ZeroMQPort:     "5555",
		ZeroMQAddr:     "localhost:5555",
	}

	assert.Equal(t, "producer", cfg.Mode)
	assert.Equal(t, "50051", cfg.GRPCPort)
	assert.Equal(t, "localhost:50051", cfg.GRPCAddr)
	assert.Equal(t, "9090", cfg.PrometheusPort)
	assert.Equal(t, "postgres://user:password@localhost:5432/tasks?sslmode=disable", cfg.DBConnString)
	assert.Equal(t, 1000, cfg.MaxBacklog)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "json", cfg.LogFormat)
	assert.Equal(t, "6060", cfg.ProfilingPort)
	assert.Equal(t, 10, cfg.MessageRate)
	assert.Equal(t, "5555", cfg.ZeroMQPort)
	assert.Equal(t, "localhost:5555", cfg.ZeroMQAddr)
}

func TestGetEnvAsInt_InvalidValues(t *testing.T) {
	testCases := []struct {
		name     string
		envValue string
		fallback int
		expected int
	}{
		{"empty string", "", 42, 42},
		{"non-numeric", "abc", 42, 42},
		{"partial numeric", "123abc", 42, 42},
		{"negative", "-123", -1, -123},
		{"zero", "0", 42, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("TEST_INT", tc.envValue)
			defer os.Unsetenv("TEST_INT")

			result := getEnvAsInt("TEST_INT", tc.fallback)
			assert.Equal(t, tc.expected, result)
		})
	}
}
