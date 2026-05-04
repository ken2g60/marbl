package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	Mode           string // "producer" or "consumer"
	GRPCPort       string
	GRPCAddr       string
	PrometheusPort string
	DBConnString   string
	MaxBacklog     int
	LogLevel       string
	LogFormat      string
	ProfilingPort  string
	MessageRate    int    // msgs/sec for producer, limit for consumer
	ZeroMQPort     string // ZeroMQ port for pub/sub
	ZeroMQAddr     string // ZeroMQ address for subscriber
}

func LoadConfig() Config {
	cfg := Config{}

	flag.StringVar(&cfg.GRPCPort, "grpc-port", getEnv("GRPC_PORT", "50051"), "gRPC port (for server)")
	flag.StringVar(&cfg.GRPCAddr, "grpc-addr", getEnv("GRPC_ADDR", "localhost:50051"), "gRPC address (for client)")
	flag.StringVar(&cfg.PrometheusPort, "prometheus-port", getEnv("PROMETHEUS_PORT", "9090"), "Prometheus metrics port")
	flag.StringVar(&cfg.DBConnString, "db-conn", getEnv("DB_CONN", "postgres://user:password@localhost:5432/tasks?sslmode=disable"), "Database connection string")
	flag.IntVar(&cfg.MaxBacklog, "max-backlog", getEnvAsInt("MAX_BACKLOG", 1000), "Maximum backlog for producer")
	flag.StringVar(&cfg.LogLevel, "log-level", getEnv("LOG_LEVEL", "info"), "Log level (debug, info, warn, error)")
	flag.StringVar(&cfg.LogFormat, "log-format", getEnv("LOG_FORMAT", "json"), "Log format (console, json)")
	flag.StringVar(&cfg.ProfilingPort, "profiling-port", getEnv("PROFILING_PORT", "6060"), "Profiling port")
	flag.IntVar(&cfg.MessageRate, "message-rate", getEnvAsInt("MESSAGE_RATE", 10), "Message rate per second")
	flag.StringVar(&cfg.ZeroMQPort, "zeromq-port", getEnv("ZEROMQ_PORT", "5555"), "ZeroMQ port for pub/sub")
	flag.StringVar(&cfg.ZeroMQAddr, "zeromq-addr", getEnv("ZEROMQ_ADDR", "localhost:5555"), "ZeroMQ address for subscriber")

	versionFlag := flag.Bool("version", false, "Print version and exit")

	flag.Parse()

	if *versionFlag {
		return Config{Mode: "version"}
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}
