package task_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/anon-org/developing-api-services-with-golang/domain"
	"github.com/anon-org/developing-api-services-with-golang/task"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	db, _ = sql.Open("sqlite3", ":memory:")
	api   = task.Wire(db)
)

func TestMain(m *testing.M) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS tasks(
	id TEXT PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	last_modified_at TIMESTAMP NOT NULL DEFAULT 0,
	is_active BOOL NOT NULL DEFAULT TRUE)`)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	m.Run()
}

func NewTaskStoreRequest(t *testing.T, name string) domain.TaskStoreRequest {
	t.Helper()
	return domain.TaskStoreRequest{
		Name: name,
	}
}

func v1TransportHTTP_Store(name string) func(t *testing.T) string {
	return func(t *testing.T) string {
		var b bytes.Buffer
		if err := json.NewEncoder(&b).Encode(NewTaskStoreRequest(t, name)); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, task.V1HTTPEndpoint, &b)
		res := httptest.NewRecorder()

		api.Route().ServeHTTP(res, req)

		if ct := res.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type to be application/json, got %s", ct)
		}
		if res.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", res.Code)
		}

		var task domain.TaskResponse
		if err := json.NewDecoder(res.Body).Decode(&task); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if task.ID == "" {
			t.Errorf("expected non-empty ID, got %s", task.ID)
		}

		if task.Name != name {
			t.Errorf("expected %s, got %s", name, task.Name)
		}

		if task.CreatedAt == 0 {
			t.Errorf("expected non-zero CreatedAt, got %v", task.CreatedAt)
		}

		if task.LastModifiedAt != 0 {
			t.Errorf("expected zero LastModifiedAt, got %v", task.LastModifiedAt)
		}

		if !task.IsActive {
			t.Errorf("expected true, got %v", task.IsActive)
		}

		return task.ID
	}
}

func TestV1TransportHTTP_Store(t *testing.T) {
	v1TransportHTTP_Store("same")(t)

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(NewTaskStoreRequest(t, "same")); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, task.V1HTTPEndpoint, &b)
	res := httptest.NewRecorder()

	api.Route().ServeHTTP(res, req)

	if ct := res.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type to be application/json, got %s", ct)
	}

	if res.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", res.Code)
	}
}

func TestV1TransportHTTP_Fetch(t *testing.T) {
	v1TransportHTTP_Store("TestV1TransportHTTP_Fetch1")(t)
	v1TransportHTTP_Store("TestV1TransportHTTP_Fetch2")(t)
	v1TransportHTTP_Store("TestV1TransportHTTP_Fetch3")(t)

	req := httptest.NewRequest(http.MethodGet, task.V1HTTPEndpoint, nil)
	res := httptest.NewRecorder()

	api.Route().ServeHTTP(res, req)

	if ct := res.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type to be application/json, got %s", ct)
	}
	if res.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", res.Code)
	}

	var tasks []domain.TaskResponse
	if err := json.NewDecoder(res.Body).Decode(&tasks); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(tasks) == 0 {
		t.Errorf("expected non-empty tasks, got %v", tasks)
	}
}

func TestV1RepositorySqlite_FetchByID(t *testing.T) {
	const taskName string = "TestV1RepositorySqlite_FetchByID"
	id := v1TransportHTTP_Store(taskName)(t)

	req := httptest.NewRequest(http.MethodGet, task.V1HTTPEndpoint+id, nil)
	res := httptest.NewRecorder()

	api.Route().ServeHTTP(res, req)

	if ct := res.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type to be application/json, got %s", ct)
	}
	if res.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", res.Code)
	}

	var tr domain.TaskResponse
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if tr.ID != id {
		t.Errorf("expected %s, got %s", id, tr.ID)
	}

	if tr.Name != taskName {
		t.Errorf("expected %s, got %s", taskName, tr.Name)
	}

	if tr.CreatedAt == 0 {
		t.Errorf("expected non-zero CreatedAt, got %v", tr.CreatedAt)
	}

	if tr.LastModifiedAt != 0 {
		t.Errorf("expected zero LastModifiedAt, got %v", tr.LastModifiedAt)
	}

	if !tr.IsActive {
		t.Errorf("expected true, got %v", tr.IsActive)
	}
}

func TestV1RepositorySqlite_Patch(t *testing.T) {
	const taskName string = "TestV1RepositorySqlite_Patch"
	id := v1TransportHTTP_Store(taskName)(t)

	t.Run("patch name only", func(t *testing.T) {
		patchName := "TestV1RepositorySqlite_Patch_PatchName"

		var b bytes.Buffer
		p := &domain.TaskPatchRequest{
			Name: &patchName,
		}

		if err := json.NewEncoder(&b).Encode(p); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		req := httptest.NewRequest(http.MethodPatch, task.V1HTTPEndpoint+id, &b)
		res := httptest.NewRecorder()

		api.Route().ServeHTTP(res, req)

		if ct := res.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type to be application/json, got %s", ct)
		}

		if res.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", res.Code)
		}

		var tr domain.TaskResponse
		if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if tr.ID != id {
			t.Errorf("expected %s, got %s", id, tr.ID)
		}

		if tr.Name != patchName {
			t.Errorf("expected %s, got %s", patchName, tr.Name)
		}

		if tr.CreatedAt == 0 {
			t.Errorf("expected non-zero CreatedAt, got %v", tr.CreatedAt)
		}

		if tr.LastModifiedAt == 0 {
			t.Errorf("expected non-zero LastModifiedAt, got %v", tr.LastModifiedAt)
		}

		if !tr.IsActive {
			t.Errorf("expected true, got %v", tr.IsActive)
		}
	})

	t.Run("patch active only", func(t *testing.T) {
		patchActive := false

		var b bytes.Buffer
		p := &domain.TaskPatchRequest{
			IsActive: &patchActive,
		}

		if err := json.NewEncoder(&b).Encode(p); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		req := httptest.NewRequest(http.MethodPatch, task.V1HTTPEndpoint+id, &b)
		res := httptest.NewRecorder()

		api.Route().ServeHTTP(res, req)

		if ct := res.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type to be application/json, got %s", ct)
		}

		if res.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", res.Code)
		}

		var tr domain.TaskResponse
		if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if tr.ID != id {
			t.Errorf("expected %s, got %s", id, tr.ID)
		}

		if tr.Name == taskName {
			t.Errorf("expected not %s, got %s", taskName, tr.Name)
		}

		if tr.CreatedAt == 0 {
			t.Errorf("expected non-zero CreatedAt, got %v", tr.CreatedAt)
		}

		if tr.LastModifiedAt == 0 {
			t.Errorf("expected non-zero LastModifiedAt, got %v", tr.LastModifiedAt)
		}

		if tr.IsActive {
			t.Errorf("expected false, got %v", tr.IsActive)
		}
	})

	t.Run("patch all", func(t *testing.T) {
		patchActive := true
		patchName := "TestV1RepositorySqlite_Patch_PatchAll"

		var b bytes.Buffer
		p := &domain.TaskPatchRequest{
			Name:     &patchName,
			IsActive: &patchActive,
		}

		if err := json.NewEncoder(&b).Encode(p); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		req := httptest.NewRequest(http.MethodPut, task.V1HTTPEndpoint+id, &b)
		res := httptest.NewRecorder()

		api.Route().ServeHTTP(res, req)

		if ct := res.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type to be application/json, got %s", ct)
		}

		if res.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", res.Code)
		}

		var tr domain.TaskResponse
		if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if tr.ID != id {
			t.Errorf("expected %s, got %s", id, tr.ID)
		}

		if tr.Name != patchName {
			t.Errorf("expected not %s, got %s", patchName, tr.Name)
		}

		if tr.CreatedAt == 0 {
			t.Errorf("expected non-zero CreatedAt, got %v", tr.CreatedAt)
		}

		if tr.LastModifiedAt == 0 {
			t.Errorf("expected non-zero LastModifiedAt, got %v", tr.LastModifiedAt)
		}

		if !tr.IsActive {
			t.Errorf("expected true, got %v", tr.IsActive)
		}
	})

	t.Run("no patch", func(t *testing.T) {
		var b bytes.Buffer
		p := &domain.TaskPatchRequest{}

		if err := json.NewEncoder(&b).Encode(p); err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		req := httptest.NewRequest(http.MethodPatch, task.V1HTTPEndpoint+id, &b)
		res := httptest.NewRecorder()

		api.Route().ServeHTTP(res, req)

		if ct := res.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type to be application/json, got %s", ct)
		}

		if res.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", res.Code)
		}
	})
}

func TestV1RepositorySqlite_DestroyByID(t *testing.T) {
	t.Run("destroy", func(t *testing.T) {
		id := v1TransportHTTP_Store("TestV1RepositorySqlite_DestroyByID")(t)

		req := httptest.NewRequest(http.MethodDelete, task.V1HTTPEndpoint+id, nil)
		res := httptest.NewRecorder()

		api.Route().ServeHTTP(res, req)

		if res.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", res.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {

		req := httptest.NewRequest(http.MethodDelete, task.V1HTTPEndpoint+"foo", nil)
		res := httptest.NewRecorder()

		api.Route().ServeHTTP(res, req)

		if res.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", res.Code)
		}
	})
}
