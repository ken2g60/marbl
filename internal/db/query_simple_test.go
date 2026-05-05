package db

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestTask(t *testing.T) {
	now := time.Now()
	task := Task{
		ID:        1,
		Type:      1,
		Value:     100,
		State:     "pending",
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	// Test Task struct
	if task.ID != 1 {
		t.Errorf("Expected ID 1, got %d", task.ID)
	}
	if task.Type != 1 {
		t.Errorf("Expected Type 1, got %d", task.Type)
	}
	if task.Value != 100 {
		t.Errorf("Expected Value 100, got %d", task.Value)
	}
	if task.State != "pending" {
		t.Errorf("Expected State 'pending', got %s", task.State)
	}
	if !task.CreatedAt.Valid {
		t.Error("Expected CreatedAt to be valid")
	}
	if !task.UpdatedAt.Valid {
		t.Error("Expected UpdatedAt to be valid")
	}
}

func TestCreateTaskParams(t *testing.T) {
	params := CreateTaskParams{
		Type:  2,
		Value: 200,
		State: "processing",
	}

	if params.Type != 2 {
		t.Errorf("Expected Type 2, got %d", params.Type)
	}
	if params.Value != 200 {
		t.Errorf("Expected Value 200, got %d", params.Value)
	}
	if params.State != "processing" {
		t.Errorf("Expected State 'processing', got %s", params.State)
	}
}

func TestUpdateTaskStateParams(t *testing.T) {
	params := UpdateTaskStateParams{
		ID:    123,
		State: "completed",
	}

	if params.ID != 123 {
		t.Errorf("Expected ID 123, got %d", params.ID)
	}
	if params.State != "completed" {
		t.Errorf("Expected State 'completed', got %s", params.State)
	}
}

func TestQueriesImplementsQuerier(t *testing.T) {
	// Test that Queries implements Querier interface
	var _ Querier = (*Queries)(nil)
}

func TestDBTXInterface(t *testing.T) {
	// Test that DBTX interface has the correct methods
	var _ DBTX = (*MockDBTX)(nil)
}
