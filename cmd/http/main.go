package main

import (
	"context"
	"database/sql"
	"github.com/anon-org/developing-api-services-with-golang/task"
	"github.com/anon-org/developing-api-services-with-golang/util/logutil"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
)

const (
	dbFileName         = "production.db.out"
	appPort            = ":8080"
	querySqliteMigrate = `CREATE TABLE IF NOT EXISTS tasks(
	id TEXT PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	last_modified_at TIMESTAMP NOT NULL DEFAULT 0,
	is_active BOOL NOT NULL DEFAULT TRUE)`
)

var (
	logger *log.Logger = logutil.NewStdLogger()
)

func migrate(ctx context.Context, db *sql.DB) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, querySqliteMigrate)
	if err != nil {
		logger.Fatal(err)
	}
}

func main() {
	db, err := sql.Open("sqlite3", dbFileName)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	migrate(context.Background(), db)
	api := task.Wire(db)

	http.Handle("/", http.FileServer(http.Dir("./public")))
	http.HandleFunc(task.V1HTTPEndpoint, api.Route())

	logger.Println("listening on", appPort)
	if err := http.ListenAndServe(appPort, nil); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
