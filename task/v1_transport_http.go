package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/anon-org/developing-api-services-with-golang/domain"
	"github.com/anon-org/developing-api-services-with-golang/util/logutil"
	"net/http"
	"strings"
	"time"
)

var (
	ErrInvalidPath error = errors.New("invalid path")
)

const (
	// V1HTTPEndpoint is the endpoint for the v1 HTTP API.
	V1HTTPEndpoint string = "/v1/tasks/"
)

type v1TransportHTTP struct {
	svc domain.TaskService
}

func (v v1TransportHTTP) Route() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// create contextual logger
		l := logutil.NewCtxLogger()
		ctx := logutil.PutCtxLogger(r.Context(), l)
		r = r.WithContext(ctx)

		// track latency
		now := time.Now()
		defer l.Println(r.Method, r.URL.Path, time.Since(now))

		switch r.Method {
		case http.MethodGet:
			if V1HTTPEndpoint == r.URL.Path {
				v.Fetch().ServeHTTP(w, r)
			} else {
				v.FetchByID().ServeHTTP(w, r)
			}
		case http.MethodPost:
			v.Store().ServeHTTP(w, r)
		case http.MethodPatch, http.MethodPut:
			v.Patch().ServeHTTP(w, r)
		case http.MethodDelete:
			v.DestroyByID().ServeHTTP(w, r)
		}
	}
}

func (v v1TransportHTTP) Fetch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logutil.GetCtxLogger(r.Context())
		w.Header().Set("Content-Type", "application/json")

		tasks, err := v.svc.Fetch(r.Context())
		if err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
			return
		}

		responses := make([]*domain.TaskResponse, len(tasks))
		for i, t := range tasks {
			responses[i] = t.ToResponse()
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(responses); err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
		}
	}
}

func (v v1TransportHTTP) FetchByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logutil.GetCtxLogger(r.Context())

		w.Header().Set("Content-Type", "application/json")

		id, err := v.extractID(r.URL.Path)
		if err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
			return
		}

		task, err := v.svc.FetchByID(r.Context(), id)
		if err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(task.ToResponse()); err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
		}
	}
}

func (v v1TransportHTTP) Store() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logutil.GetCtxLogger(r.Context())

		w.Header().Set("Content-Type", "application/json")

		var t domain.TaskStoreRequest
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
			return
		}

		stored, err := v.svc.Store(r.Context(), t.Name)
		if err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
			return
		}

		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(stored.ToResponse()); err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
		}
	}
}

func (v v1TransportHTTP) Patch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logutil.GetCtxLogger(r.Context())

		w.Header().Set("Content-Type", "application/json")

		id, err := v.extractID(r.URL.Path)
		if err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
			return
		}

		var tr domain.TaskPatchRequest
		if err := json.NewDecoder(r.Body).Decode(&tr); err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
			return
		}

		t := domain.TaskPatchSpec{
			ID:       id,
			Name:     tr.Name,
			IsActive: tr.IsActive,
		}

		patched, err := v.svc.Patch(r.Context(), t)

		if err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(patched.ToResponse()); err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
		}
	}
}

func (v v1TransportHTTP) DestroyByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logutil.GetCtxLogger(r.Context())
		w.Header().Set("Content-Type", "application/json")

		id, err := v.extractID(r.URL.Path)
		if err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
			return
		}

		if err := v.svc.DestroyByID(r.Context(), id); err != nil {
			l.Println(err)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"error": "%s"}`, err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (v v1TransportHTTP) extractID(path string) (string, error) {
	if V1HTTPEndpoint == path {
		return "", ErrInvalidPath
	}

	list := strings.Split(path, V1HTTPEndpoint)

	if len(list) < 2 {
		return "", ErrInvalidPath
	}

	return list[1], nil
}
