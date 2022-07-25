package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/anon-org/developing-api-services-with-golang/domain"
	"github.com/anon-org/developing-api-services-with-golang/util/logutil"
	"time"
)

const (
	queryDefaultTimeout time.Duration = 10 * time.Second

	querySqliteFetch = `SELECT *
FROM tasks
ORDER BY created_at ASC`

	querySqliteFetchByID = `SELECT *
FROM tasks
WHERE id = $1
LIMIT 1`

	querySqliteStore = `INSERT INTO tasks (id, name)
VALUES ($1, $2)
RETURNING *`

	querySqliteDestroy = `DELETE FROM tasks WHERE id = $1`
)

type v1RepositorySqlite struct {
	db *sql.DB
}

func (v v1RepositorySqlite) Fetch(ctx context.Context) ([]*domain.TaskEntity, error) {
	l := logutil.GetCtxLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, queryDefaultTimeout)
	defer cancel()

	rows, err := v.db.QueryContext(ctx, querySqliteFetch)
	if err != nil {
		l.Println(err)
		return nil, fmt.Errorf("%w: failed to fetch tasks", err)
	}
	defer rows.Close()

	entities := make([]*domain.TaskEntity, 0)
	for rows.Next() {
		var e domain.TaskEntity
		if err := rows.Scan(&e.ID, &e.Name, &e.CreatedAt, &e.LastModifiedAt, &e.IsActive); err != nil {
			l.Println(err)
			return nil, fmt.Errorf("%w: failed to scan tasks", err)
		}

		entities = append(entities, &e)
	}

	return entities, nil
}

func (v v1RepositorySqlite) FetchByID(ctx context.Context, id string) (*domain.TaskEntity, error) {
	l := logutil.GetCtxLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, queryDefaultTimeout)
	defer cancel()

	rows, err := v.db.QueryContext(ctx, querySqliteFetchByID, id)
	if err != nil {
		l.Println(err)
		return nil, fmt.Errorf("%w: failed to fetch task by id: %s", err, id)
	}
	defer rows.Close()

	if !rows.Next() {
		err := fmt.Errorf("task with id: %s not found", id)
		l.Println(err)
		return nil, err
	}

	var e domain.TaskEntity
	if err := rows.Scan(&e.ID, &e.Name, &e.CreatedAt, &e.LastModifiedAt, &e.IsActive); err != nil {
		l.Println(err)
		return nil, fmt.Errorf("%w: failed to scan task with id: %s", err, id)
	}

	return &e, nil
}

func (v v1RepositorySqlite) Store(ctx context.Context, entity domain.TaskEntity) (*domain.TaskEntity, error) {
	l := logutil.GetCtxLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, queryDefaultTimeout)
	defer cancel()

	rows, err := v.db.QueryContext(ctx, querySqliteStore, entity.ID, entity.Name)
	if err != nil {
		l.Println(err)
		return nil, fmt.Errorf("%w: failed to store task: %s", err, entity.Name)
	}
	defer rows.Close()

	if !rows.Next() {
		err := fmt.Errorf("failed to store task: %s", entity.Name)
		l.Println(err)
		return nil, err
	}

	if err := rows.Scan(&entity.ID, &entity.Name, &entity.CreatedAt, &entity.LastModifiedAt, &entity.IsActive); err != nil {
		l.Println(err)
		return nil, fmt.Errorf("%w: failed to scan task: %v", err, entity)
	}

	return &entity, nil
}

func (v v1RepositorySqlite) Patch(ctx context.Context, entity domain.TaskPatchSpec) (*domain.TaskEntity, error) {
	l := logutil.GetCtxLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, queryDefaultTimeout)
	defer cancel()

	querySqlitePatch, args := v.constructQuerySqlitePatch(entity)
	if len(args) <= 1 {
		return nil, errors.New("no fields to patch")
	} else {
		l.Println("constructed query:", querySqlitePatch, "with args:", args)
	}

	rows, err := v.db.QueryContext(ctx, querySqlitePatch, args...)
	if err != nil {
		l.Println(err)
		return nil, fmt.Errorf("%w: failed to patch task: %s", err, entity.ID)
	}
	defer rows.Close()

	if !rows.Next() {
		err := fmt.Errorf("failed to patch task: %s", entity.ID)
		l.Println(err)
		return nil, err
	}

	var e domain.TaskEntity
	if err := rows.Scan(&e.ID, &e.Name, &e.CreatedAt, &e.LastModifiedAt, &e.IsActive); err != nil {
		l.Println(err)
		return nil, fmt.Errorf("%w: failed to scan task: %v", err, e)
	}

	return &e, nil
}

func (v v1RepositorySqlite) DestroyByID(ctx context.Context, id string) error {
	l := logutil.GetCtxLogger(ctx)
	ctx, cancel := context.WithTimeout(ctx, queryDefaultTimeout)
	defer cancel()

	res, err := v.db.ExecContext(ctx, querySqliteDestroy, id)
	if err != nil {
		l.Println(err)
		return fmt.Errorf("%w: failed to destroy task by id: %s", err, id)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		l.Println(err)
		return fmt.Errorf("%w: failed to destroy task by id: %s", err, id)
	}

	if rowsAffected == 0 {
		err := fmt.Errorf("task with id: %s not found", id)
		l.Println(err)
		return err
	}

	return nil
}

func (v v1RepositorySqlite) constructQuerySqlitePatch(entity domain.TaskPatchSpec) (string, []any) {
	args := make([]any, 0)
	baseQuery := `UPDATE tasks
SET last_modified_at = CURRENT_TIMESTAMP`

	if entity.Name != nil {
		baseQuery = fmt.Sprintf("%s, name = ?", baseQuery)
		args = append(args, *entity.Name)
	}

	if entity.IsActive != nil {
		baseQuery = fmt.Sprintf("%s, is_active = ?", baseQuery)
		args = append(args, *entity.IsActive)
	}

	return fmt.Sprintf("%s WHERE id = ? RETURNING *", baseQuery), append(args, entity.ID)
}
