package task

import (
	"database/sql"
	"github.com/anon-org/developing-api-services-with-golang/domain"
	"sync"
)

var (
	v1RepoSqlite     *v1RepositorySqlite
	v1RepoSqliteOnce sync.Once

	v1Svc     *v1Service
	v1SvcOnce sync.Once

	v1TrpHTTP     *v1TransportHTTP
	v1TrpHTTPOnce sync.Once
)

// ProvideV1RepositorySqlite provides a v1RepositorySqlite implementation.
func ProvideV1RepositorySqlite(db *sql.DB) *v1RepositorySqlite {
	v1RepoSqliteOnce.Do(func() {
		v1RepoSqlite = &v1RepositorySqlite{
			db: db,
		}
	})

	return v1RepoSqlite
}

// ProvideV1Service provides a v1Service implementation.
func ProvideV1Service(repo domain.TaskRepository) *v1Service {
	v1SvcOnce.Do(func() {
		v1Svc = &v1Service{
			repo: repo,
		}
	})

	return v1Svc
}

// ProvideV1TransportHTTP provides a v1TransportHTTP implementation.
func ProvideV1TransportHTTP(svc domain.TaskService) *v1TransportHTTP {
	v1TrpHTTPOnce.Do(func() {
		v1TrpHTTP = &v1TransportHTTP{
			svc: svc,
		}
	})

	return v1TrpHTTP
}

// Wire provides a v1TransportHTTP implementation.
func Wire(db *sql.DB) *v1TransportHTTP {
	repo := ProvideV1RepositorySqlite(db)
	svc := ProvideV1Service(repo)
	return ProvideV1TransportHTTP(svc)
}
