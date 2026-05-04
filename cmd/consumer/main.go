package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kense/home-task/internal/config"
	"github.com/kense/home-task/internal/db"
	"github.com/kense/home-task/internal/logger"
	pb "github.com/kense/home-task/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
)

var (
	Version = "dev"

	tasksProcessedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tasks_processed_total",
		Help: "The total number of tasks started processing",
	})
	tasksDoneTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tasks_done_total",
		Help: "The total number of tasks completely done",
	})
	tasksPerType = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tasks_per_type_total",
		Help: "The total number of tasks done per task type",
	}, []string{"task_type"})
	taskValueSum = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "task_value_sum_total",
		Help: "The total sum of value field for each task type",
	}, []string{"task_type"})
)

// TaskMessage represents a task message received via TCP pub/sub
type TaskMessage struct {
	ID    int64  `json:"id"`
	Type  int32  `json:"type"`
	Value int32  `json:"value"`
	State string `json:"state"`
}

type server struct {
	pb.UnimplementedTaskServiceServer
	queries *db.Queries
	limiter *rate.Limiter

	// Track sum of value per type
	mu       sync.Mutex
	typeSums map[int32]int64
}

func (s *server) ProcessTask(ctx context.Context, req *pb.TaskRequest) (*pb.TaskResponse, error) {
	// Rate Limiting
	if err := s.limiter.Wait(ctx); err != nil {
		slog.Error("Rate limiter rejected request", "error", err)
		return nil, err
	}

	tasksProcessedTotal.Inc()

	// Update DB to "processing"
	err := s.queries.UpdateTaskState(ctx, db.UpdateTaskStateParams{
		ID:    req.Id,
		State: "processing",
	})
	if err != nil {
		slog.Error("Failed to update task to processing", "id", req.Id, "error", err)
		return nil, err
	}

	// Sleep according to task value (ms)
	time.Sleep(time.Duration(req.Value) * time.Millisecond)

	// Update DB to "done"
	err = s.queries.UpdateTaskState(ctx, db.UpdateTaskStateParams{
		ID:    req.Id,
		State: "done",
	})
	if err != nil {
		slog.Error("Failed to update task to done", "id", req.Id, "error", err)
		return nil, err
	}

	// Aggregations
	tasksDoneTotal.Inc()
	typeStr := fmt.Sprintf("%d", req.Type)
	tasksPerType.WithLabelValues(typeStr).Inc()
	taskValueSum.WithLabelValues(typeStr).Add(float64(req.Value))

	// In-memory sum calculation for final log
	s.mu.Lock()
	s.typeSums[req.Type] += int64(req.Value)
	totalSumForType := s.typeSums[req.Type]
	s.mu.Unlock()

	slog.Info("Task finished",
		"id", req.Id,
		"type", req.Type,
		"value", req.Value,
		"total_sum_for_type", totalSumForType,
	)

	return &pb.TaskResponse{Success: true}, nil
}

func main() {
	cfg := config.LoadConfig()
	if cfg.Mode == "version" {
		fmt.Printf("Task Consumer version %s\n", Version)
		os.Exit(0)
	}

	logger.SetupLogger(cfg.LogLevel, cfg.LogFormat)
	slog.Info("Starting Task Consumer", "version", Version)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Profiling
	go func() {
		slog.Info("Starting pprof", "port", cfg.ProfilingPort)
		if err := http.ListenAndServe(":"+cfg.ProfilingPort, nil); err != nil {
			slog.Error("pprof server failed", "error", err)
		}
	}()

	// Prometheus Metrics
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		slog.Info("Starting Prometheus metrics", "port", cfg.PrometheusPort)
		if err := http.ListenAndServe(":"+cfg.PrometheusPort, mux); err != nil {
			slog.Error("Metrics server failed", "error", err)
		}
	}()

	// Database
	queries, conn, err := db.ConnectAndMigrate(ctx, cfg.DBConnString)
	if err != nil {
		slog.Error("Database connection/migration failed", "error", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// gRPC Server setup
	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		slog.Error("Failed to listen on gRPC port", "port", cfg.GRPCPort, "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()

	// Create consumer service with rate limiter
	limiter := rate.NewLimiter(rate.Limit(cfg.MessageRate), cfg.MessageRate)
	srv := &server{
		queries:  queries,
		limiter:  limiter,
		typeSums: make(map[int32]int64),
	}

	pb.RegisterTaskServiceServer(grpcServer, srv)

	slog.Info("Starting gRPC Server", "port", cfg.GRPCPort)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			slog.Error("gRPC server failed", "error", err)
		}
	}()

	// TCP Subscriber for pub/sub
	go func() {
		// Retry connection to producer
		for {
			conn, err := net.Dial("tcp", cfg.ZeroMQAddr)
			if err != nil {
				slog.Error("Failed to connect to producer, retrying...", "address", cfg.ZeroMQAddr, "error", err)
				time.Sleep(5 * time.Second)
				continue
			}

			slog.Info("Connected to producer TCP pub/sub", "address", cfg.ZeroMQAddr)

			scanner := bufio.NewScanner(conn)
			for scanner.Scan() {
				var taskMsg TaskMessage
				if err := json.Unmarshal([]byte(scanner.Text()), &taskMsg); err != nil {
					slog.Error("Failed to unmarshal task message", "error", err)
					continue
				}

				slog.Debug("Received task via TCP pub/sub", "id", taskMsg.ID, "type", taskMsg.Type, "value", taskMsg.Value, "state", taskMsg.State)

				// Here you can process the received task message
				// For now, just log it - the actual processing happens via gRPC
			}

			if err := scanner.Err(); err != nil {
				slog.Error("TCP subscriber scanner error", "error", err)
			}

			conn.Close()
			slog.Info("TCP subscriber disconnected, retrying...")
			time.Sleep(5 * time.Second)
		}
	}()

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	slog.Info("Received termination signal, shutting down gracefully...")
	grpcServer.GracefulStop()
}
