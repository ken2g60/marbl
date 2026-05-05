package benchmark

import (
	"context"
	"testing"
	"time"

	"github.com/kense/home-task/internal/config"
	"github.com/kense/home-task/internal/db"
	"github.com/kense/home-task/internal/logger"
)

func BenchmarkConfigLoading(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Clear flags between runs
		b.ResetTimer()
		_ = config.LoadConfig()
	}
}

func BenchmarkLoggerSetup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.ResetTimer()
		log := logger.SetupLogger("info", "json")
		_ = log
	}
}

func BenchmarkDatabaseConnection(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark in short mode")
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test connection failure (this is what we can benchmark without a real DB)
		_, _, err := db.ConnectAndMigrate(ctx, "invalid://connection")
		if err == nil {
			b.Fatal("Expected error for invalid connection")
		}
	}
}

func BenchmarkLoggerOperations(b *testing.B) {
	log := logger.SetupLogger("info", "json")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		log.Info("benchmark test message", "iteration", i)
	}
}

func BenchmarkConfigEnvParsing(b *testing.B) {
	// Test environment variable parsing performance
	for i := 0; i < b.N; i++ {
		b.ResetTimer()
		_ = config.LoadConfig()
	}
}

// Note: These benchmarks would be more meaningful with:
// 1. Real database operations
// 2. gRPC service calls
// 3. Message processing
// 4. Concurrent operations

func BenchmarkConcurrentLogger(b *testing.B) {
	log := logger.SetupLogger("info", "json")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Info("concurrent benchmark test")
		}
	})
}

func BenchmarkTimeOperations(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = time.Now()
		_ = time.Since(time.Now())
	}
}
