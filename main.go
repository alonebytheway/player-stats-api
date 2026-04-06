package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/pressly/goose"
)

type PlayerService struct {
	repo Repository
}

type PlayerHandler struct {
	service *PlayerService
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Fill .env is not found")
	}
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

	go backgroundWorker()

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Server has started by port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
