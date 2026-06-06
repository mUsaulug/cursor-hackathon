package store

import (
	"context"
	"sync"

	task "cursor-hackathon/backend/internal/domain/task"
)

// TaskInMemory is a goroutine-safe in-memory task store.
type TaskInMemory struct {
	mu    sync.RWMutex
	byID  map[string]task.Task
	order []string
}

// NewTaskInMemory builds the store.
func NewTaskInMemory() *TaskInMemory {
	return &TaskInMemory{byID: map[string]task.Task{}}
}

// Save stores or replaces a task.
func (s *TaskInMemory) Save(_ context.Context, t task.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byID[t.TaskID]; !ok {
		s.order = append(s.order, t.TaskID)
	}
	s.byID[t.TaskID] = t
	return nil
}

// Get returns a task by id.
func (s *TaskInMemory) Get(_ context.Context, id string) (task.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.byID[id]
	if !ok {
		return task.Task{}, task.ErrTaskNotFound
	}
	return t, nil
}

// List returns tasks in insertion order.
func (s *TaskInMemory) List(_ context.Context) []task.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]task.Task, 0, len(s.order))
	for _, id := range s.order {
		out = append(out, s.byID[id])
	}
	return out
}
