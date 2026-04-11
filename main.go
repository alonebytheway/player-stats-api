package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/pressly/goose"

	_ "player-stats-api/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

type PlayerService struct {
	repo Repository
}

type PlayerHandler struct {
	service *PlayerService
	cache   *LeaderboardCache
}

// @title Player Stats API
// @version 1.0
// @description API сервис для отслеживания статистики игроков в дуэлях.
// @host localhost:8081
// @BasePath /
//
// @securityDefinitions.apikey secret
// @in header
// @name Authorization
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGABRT)
	defer stop()

	repo := &PlayerRepository{db: db}
	service := &PlayerService{repo: repo}

	leaderboardCache := &LeaderboardCache{}
	topPlayerChannel := make(chan []LeaderboardEntry)

	go fetchLeaderboardWorker(ctx, service, topPlayerChannel)
	go updateCacheWorker(leaderboardCache, topPlayerChannel)

	handler := &PlayerHandler{
		service: service,
		cache:   leaderboardCache,
	}

	r := chi.NewRouter()

	r.Use(RecovererMiddleware)
	r.Use(LoggingMiddleware)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware)

		r.Get("/leaderboard", handler.GetLeaderboard)

		r.Route("/players", func(r chi.Router) {
			r.Get("/", handler.GetPlayers)
			r.Post("/", handler.CreatePlayer)
			r.Delete("/{name}", handler.DeletePlayer)
			r.Patch("/{name}", handler.UpdatePlayer)
			r.Post("/duel", handler.RecordDuel)
		})
	})
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server Error", err)
		}
	}()
	log.Println("Server was Started")
	<-ctx.Done()
	log.Println("Take a signal, stop the server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Stopping Error")
	}
	log.Println("Server succsesful stoped")
}
