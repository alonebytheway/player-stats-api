package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	_ "github.com/lib/pq"

	"time"
)

var db *sql.DB

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fmt.Println(start, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		fmt.Println(r.Method, r.URL.Path, "-", duration)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "secret" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)

	})
}

func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}
		return handler
	}
}

func (h *PlayerHandler) HandlerLeaderboard(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	leaderboard, err := h.service.GetLeaderboard()

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(leaderboard)

}

func (h *PlayerHandler) HandlePlayers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		players, err := h.service.GetAll()

		if err != nil {
			http.Error(w, "Service error", http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(players)

	case http.MethodPost:
		var p Player

		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		err := h.service.CreatePlayer(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *PlayerHandler) HandlePlayerByName(w http.ResponseWriter, r *http.Request) {

	name := strings.TrimPrefix(r.URL.Path, "/players/")

	if name == "" {
		http.Error(w, "Player name is required", http.StatusBadRequest)
		return
	}

	switch r.Method {

	case http.MethodGet:

		for _, p := range players {
			if p.Name == name {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(p)
				return
			}
		}

		http.Error(w, "Player not found", http.StatusNotFound)

	case http.MethodDelete:
		name := strings.TrimPrefix(r.URL.Path, "/player/")
		if name == "" {
			http.Error(w, "Player name is required", http.StatusBadRequest)
			return
		}

		err := h.service.Delete(name)

		if err != nil {
			switch err {
			case ErrorBadRequest:
				http.Error(w, err.Error(), http.StatusBadRequest)
			case ErrorNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, "Sever error", http.StatusInternalServerError)
			}
			return
		}

		json.NewEncoder(w).Encode(Response{
			Status: "deleted",
		})

	case http.MethodPatch:

		var update UpdatePlayer

		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		name := strings.TrimPrefix(r.URL.Path, "/players/")
		if name == "" {
			http.Error(w, "Player name is required", http.StatusBadRequest)
			return
		}

		err := h.service.Update(name, update)

		if err != nil {
			switch err {
			case ErrorBadRequest:
				http.Error(w, err.Error(), http.StatusBadRequest)
			case ErrorNotFound:
				http.Error(w, err.Error(), http.StatusNotFound)
			default:
				http.Error(w, "Sever error", http.StatusInternalServerError)
			}
			return
		}

		json.NewEncoder(w).Encode(Response{
			Status: "patched",
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
