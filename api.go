package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type APIFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type APIServer struct {
	router        *mux.Router
	listenAddress string
}

func NewAPIServer() *APIServer {
	return &APIServer{
		router: mux.NewRouter(),
	}
}

func (s *APIServer) Run() error {

	s.router.HandleFunc("/", handleFunc(s.index))

	return http.ListenAndServe(s.listenAddress, nil)
}

func handleFunc(fn APIFunc) http.HandlerFunc {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "request_id", uuid.New())
	return func(w http.ResponseWriter, r *http.Request) {
		err := fn(ctx, w, r)
		if err != nil {
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
	}
}

func (s *APIServer) index(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	return nil
}
