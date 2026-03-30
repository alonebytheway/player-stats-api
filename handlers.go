package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"

	"time"

	"github.com/go-chi/chi/v5"
)

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

func (h *PlayerHandler) DeletePlayer(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	err := h.service.DeletePlayer(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PlayerHandler) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var update UpdatePlayer
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	name := chi.URLParam(r, "name")
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	err := h.service.Update(name, update)
	if err != nil {
		if errors.Is(err, ErrorNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{Status: "patched"})
}

func (h *PlayerHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {

	leaderboard, err := h.service.GetLeaderboard()

	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)

}
func (h *PlayerHandler) GetPlayers(w http.ResponseWriter, r *http.Request) {
	players, err := h.service.GetAll()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(players)
}

func (h *PlayerHandler) CreatePlayer(w http.ResponseWriter, r *http.Request) {
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
}

func (h *PlayerHandler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	player, err := h.service.GetPlayer(name)
	if err != nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}
