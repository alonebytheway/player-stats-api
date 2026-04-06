package main

import (
	"net/http"
)

type Player struct {
	Name    string `json:"name"`
	Kills   int    `json:"kills"`
	Deaths  int    `json:"deaths"`
	Matches int    `json:"matches"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

type LeaderboardEntry struct {
	Name string  `json:"name"`
	KD   float64 `json:"kd"`
}

type UpdatePlayer struct {
	Kills   *int `json:"kills"`
	Deaths  *int `json:"deaths"`
	Matches *int `json:"matches"`
}

type TopPlayer struct {
	Limit int `json:"limit"`
}

type contextKey string

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

const userContextKey contextKey = "userID"

type DuelRequest struct {
	Winner string `json:"winner"`
	Loser  string `json:"loser"`
}
