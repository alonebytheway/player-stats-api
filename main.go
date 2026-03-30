package main

import (
	"net/http"
)

type PlayerService struct {
	repo *PlayerRepository
}

type PlayerHandler struct {
	service *PlayerService
}

func main() {

	db := ConnectDB()
	defer db.Close()

	repo := &PlayerRepository{db: db}
	service := &PlayerService{repo: repo}
	handler := &PlayerHandler{service: service}

	chain := Chain(LoggingMiddleware, AuthMiddleware)

	http.Handle("/players", chain(http.HandlerFunc(handler.HandlePlayers)))

	http.Handle("/leaderboard", chain(http.HandlerFunc(handler.GetLeaderboard)))

	http.Handle("/players/", chain(http.HandlerFunc(handler.HandlePlayerByName)))

	http.ListenAndServe(":8080", nil)

}
