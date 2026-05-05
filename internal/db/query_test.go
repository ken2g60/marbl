package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockQueries extends Queries for testing
type MockQueries struct {
	*Queries
	tasks []Task
}

func NewMockQueries() *MockQueries {
	return &MockQueries{
		tasks: make([]Task, 0),
	}
}

func (m *MockQueries) CreateTask(ctx context.Context, arg CreateTaskParams) (Task, error) {
	task := Task{
		ID:        int64(len(m.tasks) + 1),
		Type:      arg.Type,
		Value:     arg.Value,
		State:     arg.State,
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	m.tasks = append(m.tasks, task)
	return task, nil
}

func (m *MockQueries) GetTasksCount(ctx context.Context) (int64, error) {
	return int64(len(m.tasks)), nil
}

func (m *MockQueries) UpdateTaskState(ctx context.Context, arg UpdateTaskStateParams) error {
	for i, task := range m.tasks {
		if task.ID == arg.ID {
			m.tasks[i].State = arg.State
			m.tasks[i].UpdatedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
			return nil
		}
	}
	return sql.ErrNoRows
}

func TestCreateTask(t *testing.T) {
	ctx := context.Background()
	mockQueries := NewMockQueries()

	params := CreateTaskParams{
		Type:  1,
		Value: 100,
		State: "pending",
	}

	task, err := mockQueries.CreateTask(ctx, params)
	require.NoError(t, err)
	assert.Equal(t, int64(1), task.ID)
	assert.Equal(t, int32(1), task.Type)
	assert.Equal(t, int32(100), task.Value)
	assert.Equal(t, "pending", task.State)
	assert.True(t, task.CreatedAt.Valid)
	assert.True(t, task.UpdatedAt.Valid)
}

func TestGetTasksCount(t *testing.T) {
	ctx := context.Background()
	mockQueries := NewMockQueries()

	// Test empty count
	count, err := mockQueries.GetTasksCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Add some tasks
	mockQueries.CreateTask(ctx, CreateTaskParams{Type: 1, Value: 100, State: "pending"})
	mockQueries.CreateTask(ctx, CreateTaskParams{Type: 2, Value: 200, State: "processing"})

	count, err = mockQueries.GetTasksCount(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestUpdateTaskState(t *testing.T) {
	ctx := context.Background()
	mockQueries := NewMockQueries()

	// Create a task first
	task, err := mockQueries.CreateTask(ctx, CreateTaskParams{Type: 1, Value: 100, State: "pending"})
	require.NoError(t, err)

	// Update the task state
	updateParams := UpdateTaskStateParams{
		ID:    task.ID,
		State: "completed",
	}

	err = mockQueries.UpdateTaskState(ctx, updateParams)
	require.NoError(t, err)

	// Verify the update
	updatedTask := mockQueries.tasks[0]
	assert.Equal(t, "completed", updatedTask.State)
	assert.Equal(t, task.ID, updatedTask.ID)
	assert.Equal(t, task.Type, updatedTask.Type)
	assert.Equal(t, task.Value, updatedTask.Value)
}

func TestUpdateTaskState_NotFound(t *testing.T) {
	ctx := context.Background()
	mockQueries := NewMockQueries()

	updateParams := UpdateTaskStateParams{
		ID:    999,
		State: "completed",
	}

	err := mockQueries.UpdateTaskState(ctx, updateParams)
	assert.Error(t, err)
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestTaskModel(t *testing.T) {
	now := time.Now()
	task := Task{
		ID:        1,
		Type:      1,
		Value:     100,
		State:     "pending",
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	assert.Equal(t, int64(1), task.ID)
	assert.Equal(t, int32(1), task.Type)
	assert.Equal(t, int32(100), task.Value)
	assert.Equal(t, "pending", task.State)
	assert.True(t, task.CreatedAt.Valid)
	assert.True(t, task.UpdatedAt.Valid)
}

func TestCreateTaskParam(t *testing.T) {
	params := CreateTaskParams{
		Type:  2,
		Value: 200,
		State: "processing",
	}

	assert.Equal(t, int32(2), params.Type)
	assert.Equal(t, int32(200), params.Value)
	assert.Equal(t, "processing", params.State)
}

func TestUpdateTaskStateParam(t *testing.T) {
	params := UpdateTaskStateParams{
		ID:    123,
		State: "completed",
	}

	assert.Equal(t, int64(123), params.ID)
	assert.Equal(t, "completed", params.State)
}
