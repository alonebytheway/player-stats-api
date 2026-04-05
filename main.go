package main

import (
	"net/http"

	"log"

	"github.com/go-chi/chi/v5"
	"github.com/pressly/goose"
)

type PlayerService struct {
	repo *PlayerRepository
}

type PlayerHandler struct {
	service *PlayerService
}

func main() {
	db := ConnectDB()
	err := goose.Up(db, "migrations")
	if err != nil {
		log.Fatalf("faild to apply migration: %v", err)
	}
	defer db.Close()

	repo := &PlayerRepository{db: db}
	service := &PlayerService{repo: repo}
	handler := &PlayerHandler{service: service}

	r := chi.NewRouter()

	r.Use(LoggingMiddleware)
	r.Use(AuthMiddleware)

	r.Get("/leaderboard", handler.GetLeaderboard)

	r.Route("/players", func(r chi.Router) {
		r.Get("/", handler.GetPlayers)

		r.Post("/", handler.CreatePlayer)
		r.Delete("/{name}", handler.DeletePlayer)
		r.Patch("/{name}", handler.UpdatePlayer)
		r.Post("/duel", handler.RecordDuel)
	})

	http.ListenAndServe(":8080", r)
}
