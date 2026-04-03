package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	_ "github.com/lib/pq"

	"time"

	"github.com/go-chi/chi/v5"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})

}

func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userContextKey).(int)
	return id, ok
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		fmt.Println(start, r.Method, r.URL.Path)
		lrw := &LoggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		fmt.Println(r.Method, r.URL.Path, "-", duration, "Status:", lrw.statusCode)
	})
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token != "secret" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID := 42
		ctx := r.Context()
		ctx = context.WithValue(ctx, userContextKey, userID)
		reqWithCtx := r.WithContext(ctx)

		next.ServeHTTP(w, reqWithCtx)

	})
}

func (h *PlayerHandler) GetPlayers(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	players, err := h.service.GetAll(ctx)
	if err != nil {
		if errors.Is(err, ErrorNotFound) {
			writeError(w, http.StatusNotFound, "Player not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Server error")
		return
	}
	writeJSON(w, http.StatusOK, players)
}

func (h *PlayerHandler) CreatePlayer(w http.ResponseWriter, r *http.Request) {

	_, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	var p Player

	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	err := h.service.CreatePlayer(ctx, p)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			writeError(w, http.StatusGatewayTimeout, "db timeout")
			return
		}

		if errors.Is(err, ErrorBadRequest) || errors.Is(err, ErrorInvalidStats) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "server error")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *PlayerHandler) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var update UpdatePlayer
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	name := chi.URLParam(r, "name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name required")
		return
	}

	err := h.service.Update(name, update)
	if err != nil {
		if errors.Is(err, ErrorNotFound) {
			writeError(w, http.StatusNotFound, "Player not found")
			return
		}
		if errors.Is(err, ErrorBadRequest) {
			writeError(w, http.StatusBadRequest, err.Error())
		}
		writeError(w, http.StatusInternalServerError, "Server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *PlayerHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {

	players, err := h.service.GetAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Server error")
		return
	}

	leaderboard, err := h.service.buildeLeaderboard(players)

	if err != nil {
		writeError(w, http.StatusInternalServerError, "Server error")
		return
	}

	writeJSON(w, http.StatusOK, leaderboard)

}

func (h *PlayerHandler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	player, err := h.service.GetPlayer(name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Server error")
		return
	}

	writeJSON(w, http.StatusOK, player)
}

func (h *PlayerHandler) DeletePlayer(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	err := h.service.DeletePlayer(name)
	if err != nil {
		if errors.Is(err, ErrorNotFound) {
			writeError(w, http.StatusNotFound, "Player not found")
			return
		}
		if errors.Is(err, ErrorBadRequest) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "Server error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PlayerHandler) GetTopPlayers(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	pageStr := r.URL.Query().Get("page")

	if limitStr == "" {
		writeError(w, http.StatusBadRequest, "limit is required")
		return
	}
	if pageStr == "" {
		writeError(w, http.StatusBadRequest, "page is required")
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		writeError(w, http.StatusBadRequest, "invalid limit")
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		writeError(w, http.StatusBadRequest, "invalid page")
		return
	}

	offset := (page - 1) * limit

	players, err := h.service.GetTopPlayers(r.Context(), limit, offset)
	if err != nil {
		if errors.Is(err, ErrorBadRequest) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "Server error")
		return
	}
	writeJSON(w, http.StatusOK, players)
}

func (h *PlayerHandler) RecordDuel(w http.ResponseWriter, r *http.Request) {
	_, ok := UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusBadRequest, "invalid user ID")
		return
	}
	var duelRequest DuelRequest
	if err := json.NewDecoder(r.Body).Decode(&duelRequest); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	err := h.service.RecordDuel(r.Context(), duelRequest.Winner, duelRequest.Loser)
	if err != nil {
		if errors.Is(err, ErrorBadRequest) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "Server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "duel recorded"})
}
