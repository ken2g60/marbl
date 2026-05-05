package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// MockDBTX implements the DBTX interface for testing
type MockDBTX struct {
	// Add mock implementation as needed
}

func (m *MockDBTX) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	// Mock implementation
	return pgconn.CommandTag{}, nil
}

func (m *MockDBTX) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	// Mock implementation
	return nil, nil
}

func (m *MockDBTX) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	// Mock implementation
	return nil
}

func TestNew(t *testing.T) {
	mockDB := &MockDBTX{}

	queries := New(mockDB)

	if queries == nil {
		t.Fatal("Expected non-nil queries")
	}

	if queries.db != mockDB {
		t.Errorf("Expected db to be mockDB, got %v", queries.db)
	}
}

func TestQueriesWithTx(t *testing.T) {
	mockDB := &MockDBTX{}
	queries := New(mockDB)

	// Note: This test would need a proper mock transaction in real implementation
	// For now, we'll just test that the method exists and returns a Queries
	newQueries := queries.WithTx(nil) // We'll pass nil for now since we don't have a real tx mock

	if newQueries == nil {
		t.Fatal("Expected non-nil newQueries")
	}
}

func TestConnectAndMigrate_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Test with invalid connection string first
	_, _, err := ConnectAndMigrate(ctx, "invalid://connection")
	if err == nil {
		t.Error("Expected error for invalid connection string")
	}

	if !contains(err.Error(), "failed to connect to db") {
		t.Errorf("Expected error to contain 'failed to connect to db', got %s", err.Error())
	}

	// Note: The actual database connection test would require setting up a test database
	// For now, we'll just test the error case
}

func TestQueriesInterface(t *testing.T) {
	mockDB := &MockDBTX{}
	queries := New(mockDB)

	// Verify that Queries implements the Querier interface
	var _ Querier = queries

	if queries == nil {
		t.Fatal("Expected non-nil queries")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
