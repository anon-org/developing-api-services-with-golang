package domain

import (
	"context"
	"time"
)

type (
	// TaskStoreRequest is the specification that represents a task HTTP Store request.
	TaskStoreRequest struct {
		Name string `json:"name"`
	}

	// TaskPatchRequest is the specification that represents a task HTTP Patch request.
	TaskPatchRequest struct {
		Name     *string `json:"name"`
		IsActive *bool   `json:"is_active"`
	}

	// TaskResponse is the specification that represents a task HTTP response.
	TaskResponse struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		CreatedAt      int64  `json:"created_at"`
		LastModifiedAt int64  `json:"last_modified_at,omitempty"`
		IsActive       bool   `json:"is_active"`
	}

	// Task is the specification that represents a task.
	Task struct {
		ID             string
		Name           string
		CreatedAt      time.Time
		LastModifiedAt time.Time
		IsActive       bool
	}

	// TaskPatchSpec is the specification that represents a task patch specification.
	TaskPatchSpec struct {
		ID       string
		Name     *string
		IsActive *bool
	}

	// TaskEntity is the repository entity that represents a task.
	TaskEntity struct {
		ID             string
		Name           string
		CreatedAt      time.Time
		LastModifiedAt time.Time
		IsActive       bool
	}

	// TaskRepository is the storage interface for TaskEntity.
	TaskRepository interface {
		Fetch(context.Context) ([]*TaskEntity, error)
		FetchByID(context.Context, string) (*TaskEntity, error)
		Store(context.Context, TaskEntity) (*TaskEntity, error)
		Patch(context.Context, TaskPatchSpec) (*TaskEntity, error)
		DestroyByID(context.Context, string) error
	}

	// TaskService is the use case interface for Task.
	TaskService interface {
		Fetch(context.Context) ([]*Task, error)
		FetchByID(context.Context, string) (*Task, error)
		Store(context.Context, string) (*Task, error)
		Patch(context.Context, TaskPatchSpec) (*Task, error)
		DestroyByID(context.Context, string) error
	}
)

// ToEntity converts a Task to a TaskEntity.
func (t *Task) ToEntity() TaskEntity {
	return TaskEntity{
		ID:             t.ID,
		Name:           t.Name,
		CreatedAt:      t.CreatedAt,
		LastModifiedAt: t.LastModifiedAt,
		IsActive:       t.IsActive,
	}
}

func (t *Task) ToResponse() *TaskResponse {
	return &TaskResponse{
		ID:             t.ID,
		Name:           t.Name,
		CreatedAt:      t.CreatedAt.UnixMilli(),
		LastModifiedAt: t.LastModifiedAt.UnixMilli(),
		IsActive:       t.IsActive,
	}
}

// ToSpec converts a TaskEntity to a Task.
func (e *TaskEntity) ToSpec() *Task {
	return &Task{
		ID:             e.ID,
		Name:           e.Name,
		CreatedAt:      e.CreatedAt,
		LastModifiedAt: e.LastModifiedAt,
		IsActive:       e.IsActive,
	}
}
