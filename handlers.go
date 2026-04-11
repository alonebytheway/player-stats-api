package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"golang.org/x/time/rate"

	_ "github.com/lib/pq"

	"time"

	"github.com/go-chi/chi/v5"
)

var visitors = make(map[string]*rate.Limiter)
var mu sync.Mutex

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

func getVisitors(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(3, 5)
		visitors[ip] = limiter
	}
	return limiter
}

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		limiter := getVisitors(ip)

		if !limiter.Allow() {
			http.Error(w, "Too Many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RecovererMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()

			if err != nil {
				fmt.Println("CRITICAL ERROR panic:", err)

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)

	})
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

// @Summary Обновить статистику игрока
// @Description Частичное обновление данных игрока (убийства, смерти, матчи). Можно передать только те поля, которые нужно изменить.
// @Tags players
// @Accept json
// @Produce json
// @Security secret
// @Param name path string true "Имя игрока"
// @Param input body UpdatePlayer true "Данные для обновления (JSON)"
// @Success 200 {object} map[string]string "Успешное обновление"
// @Failure 400 {object} map[string]string "Неверный JSON или параметры"
// @Failure 404 {object} map[string]string "Игрок не найден"
// @Failure 422 {object} map[string]string "Ошибка валидации (например, отрицательные значения)"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /players/{name} [patch]
func (h *PlayerHandler) UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var update UpdatePlayer
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&update); err != nil {
		writeError(w, http.StatusBadRequest, "Invalide JSON or unknown fields")
		return
	}
	if err := update.Validate(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	ctx := r.Context()
	name := chi.URLParam(r, "name")
	if name == "" {
		writeError(w, http.StatusBadRequest, "name required")
		return
	}

	err := h.service.Update(ctx, name, update)
	if err != nil {
		if errors.Is(err, ErrorNotFound) {
			writeError(w, http.StatusNotFound, "Player not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (u *UpdatePlayer) Validate() error {
	if u.Kills != nil && *u.Kills < 0 {
		return ErrorBadRequest
	}
	if u.Deaths != nil && *u.Deaths < 0 {
		return ErrorBadRequest
	}
	if u.Matches != nil && *u.Matches < 0 {
		return ErrorBadRequest
	}

	if u.Kills == nil && u.Deaths == nil && u.Matches == nil {
		return ErrorBadRequest
	}

	return nil
}

func (h *PlayerHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {

	h.cache.mu.RLock()

	top := h.cache.entries

	h.cache.mu.RUnlock()

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(top)

}

func (h *PlayerHandler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	ctx := r.Context()

	player, err := h.service.GetPlayer(ctx, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Server error")
		return
	}

	writeJSON(w, http.StatusOK, player)
}

func (h *PlayerHandler) DeletePlayer(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	ctx := r.Context()
	if name == "" {
		http.Error(w, "name required", http.StatusBadRequest)
		return
	}

	err := h.service.DeletePlayer(ctx, name)
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
