package task

import (
	"context"
	"fmt"
	"github.com/anon-org/developing-api-services-with-golang/domain"
	"github.com/anon-org/developing-api-services-with-golang/util/idutil"
	"github.com/anon-org/developing-api-services-with-golang/util/logutil"
)

const (
	defaultIdLength = 24
)

type v1Service struct {
	repo domain.TaskRepository
}

func (v v1Service) Fetch(ctx context.Context) ([]*domain.Task, error) {
	l := logutil.GetCtxLogger(ctx)

	entities, err := v.repo.Fetch(ctx)
	if err != nil {
		l.Println(err)
		return nil, fmt.Errorf("%w: failed to fetch tasks", err)
	}

	n := len(entities)
	tasks := make([]*domain.Task, n)
	for i, entity := range entities {
		tasks[i] = entity.ToSpec()
	}

	return tasks, nil
}

func (v v1Service) FetchByID(ctx context.Context, id string) (*domain.Task, error) {
	l := logutil.GetCtxLogger(ctx)

	entity, err := v.repo.FetchByID(ctx, id)
	if err != nil {
		l.Println(err)
		return nil, fmt.Errorf("%w: failed to fetch task by id: %s", err, id)
	}

	return entity.ToSpec(), nil
}

func (v v1Service) Store(ctx context.Context, name string) (*domain.Task, error) {
	l := logutil.GetCtxLogger(ctx)

	e := domain.TaskEntity{
		ID:   idutil.MustGenerateID(defaultIdLength),
		Name: name,
	}

	stored, err := v.repo.Store(ctx, e)
	if err != nil {
		l.Println(err)
		return nil, fmt.Errorf("%w: failed to store task: %s", err, name)
	}

	return stored.ToSpec(), nil
}

func (v v1Service) Patch(ctx context.Context, spec domain.TaskPatchSpec) (*domain.Task, error) {
	l := logutil.GetCtxLogger(ctx)

	patched, err := v.repo.Patch(ctx, spec)
	if err != nil {
		l.Println(err)
		return nil, fmt.Errorf("%w: failed to patch task: %s", err, spec.ID)
	}

	return patched.ToSpec(), nil
}

func (v v1Service) DestroyByID(ctx context.Context, id string) error {
	l := logutil.GetCtxLogger(ctx)

	err := v.repo.DestroyByID(ctx, id)
	if err != nil {
		l.Println(err)
		return fmt.Errorf("%w: failed to destroy task by id: %s", err, id)
	}

	return nil
}
