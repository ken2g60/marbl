package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
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
	"google.golang.org/grpc/credentials/insecure"
)

var (
	Version = "dev"

	tasksProducedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "tasks_produced_total",
		Help: "The total number of produced tasks",
	})
)

// TaskMessage represents a task message sent via ZeroMQ
type TaskMessage struct {
	ID    int64  `json:"id"`
	Type  int32  `json:"type"`
	Value int32  `json:"value"`
	State string `json:"state"`
}

func main() {
	cfg := config.LoadConfig()
	if cfg.Mode == "version" {
		fmt.Printf("Task Producer version %s\n", Version)
		os.Exit(0)
	}

	logger.SetupLogger(cfg.LogLevel, cfg.LogFormat)
	slog.Info("Starting Task Producer", "version", Version)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TCP Publisher for pub/sub
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.ZeroMQPort))
	if err != nil {
		slog.Error("Failed to create TCP listener", "error", err)
		os.Exit(1)
	}
	defer listener.Close()

	slog.Info("TCP publisher started", "port", cfg.ZeroMQPort)

	// Handle subscriber connections
	var subscribers []net.Conn
	var subscribersMutex sync.Mutex

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				slog.Error("Failed to accept connection", "error", err)
				continue
			}

			subscribersMutex.Lock()
			subscribers = append(subscribers, conn)
			subscribersMutex.Unlock()

			slog.Info("New subscriber connected", "remote", conn.RemoteAddr())
		}
	}()

	// Broadcast function
	broadcastTask := func(task TaskMessage) {
		data, err := json.Marshal(task)
		if err != nil {
			slog.Error("Failed to marshal task", "error", err)
			return
		}

		subscribersMutex.Lock()
		defer subscribersMutex.Unlock()

		for i, conn := range subscribers {
			_, err := conn.Write(append(data, '\n'))
			if err != nil {
				slog.Error("Failed to send to subscriber", "remote", conn.RemoteAddr(), "error", err)
				conn.Close()
				// Remove failed connection
				subscribers = append(subscribers[:i], subscribers[i+1:]...)
			}
		}
	}

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

	// gRPC Client
	grpcConn, err := grpc.Dial(cfg.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("Failed to connect to gRPC server", "error", err)
		os.Exit(1)
	}
	defer grpcConn.Close()
	client := pb.NewTaskServiceClient(grpcConn)

	slog.Info("Connected to Consumer gRPC", "addr", cfg.GRPCAddr)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Production Loop
	limiter := rate.NewLimiter(rate.Limit(cfg.MessageRate), cfg.MessageRate)

	produced := 0
	rand.Seed(time.Now().UnixNano())

	for produced < cfg.MaxBacklog {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, shutting down")
			return
		case <-sigChan:
			slog.Info("Received termination signal, shutting down")
			return
		default:
		}

		if err := limiter.Wait(ctx); err != nil {
			slog.Error("Rate limiter wait error", "error", err)
			continue
		}

		taskType := int32(rand.Intn(10))   // 0-9
		taskValue := int32(rand.Intn(100)) // 0-99

		// Create in DB
		dbTask, err := queries.CreateTask(ctx, db.CreateTaskParams{
			Type:  taskType,
			Value: taskValue,
			State: "received",
		})
		if err != nil {
			slog.Error("Failed to insert task to DB", "error", err)
			continue
		}

		// Send to Consumer
		req := &pb.TaskRequest{
			Id:    dbTask.ID,
			Type:  taskType,
			Value: taskValue,
		}

		_, err = client.ProcessTask(ctx, req)
		if err != nil {
			slog.Error("Failed to send task to consumer", "error", err)
			// Wait, if it fails, maybe retry or just move on?
			// For this task, we can just log.
		} else {
			tasksProducedTotal.Inc()
			produced++
			slog.Debug("Produced task", "id", dbTask.ID, "type", taskType, "value", taskValue)

			// Broadcast task via TCP pub/sub
			taskMsg := TaskMessage{
				ID:    dbTask.ID,
				Type:  taskType,
				Value: taskValue,
				State: "produced",
			}
			broadcastTask(taskMsg)
		}
	}

	slog.Info("Max backlog reached, stopping production", "produced", produced)
}
