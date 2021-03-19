package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const invalidName = "invalid middleware name"

type statefulMiddleware struct {
	current     func(http.Handler) http.Handler
	middlewares map[string]func(http.Handler) http.Handler
}

func (s *statefulMiddleware) update(nextName string) error {
	next, ok := s.middlewares[nextName]
	if !ok {
		return errors.New(invalidName)
	}

	s.current = next

	return nil
}

func (s *statefulMiddleware) main(next http.Handler) http.Handler {
	return s.current(next)
}

func newStatefulMiddleware(first string, middlewares map[string]func(http.Handler) http.Handler) (*statefulMiddleware, error) {
	current, ok := middlewares[first]
	if !ok {
		return nil, errors.New(invalidName)
	}

	stateful := statefulMiddleware{current: current, middlewares: middlewares}
	return &stateful, nil
}

func noVerboseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware example from https://github.com/gorilla/mux#examples
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)

		defer func(startedAt time.Time) {
			log.Println(r.RequestURI, time.Since(startedAt))
		}(time.Now())

		next.ServeHTTP(w, r)
	})
}

type service struct {
	middleware *statefulMiddleware
}

func (s *service) home(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s *service) config(w http.ResponseWriter, r *http.Request) {
	options := map[string]string{}

	err := json.NewDecoder(r.Body).Decode(&options)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	option, ok := options["option"]
	if !ok {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.middleware.update(option)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"changed": true})
}

func main() {
	stateful, _ := newStatefulMiddleware(
		"no_verbose",
		map[string]func(http.Handler) http.Handler{
			"no_verbose": noVerboseMiddleware,
			"verbose":    loggingMiddleware})

	s := service{stateful}

	r := mux.NewRouter()

	r.HandleFunc("/config", s.config)
	r.HandleFunc("/", s.home)

	r.Use(s.middleware.main)

	log.Println("starting :8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
