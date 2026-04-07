package main

import (
	"context"
	"log"
	"time"
)

func fetchLeaderboardWorker(service *PlayerService, ch chan<- []LeaderboardEntry) {
	for {
		leader, err := service.GetTopPlayers(context.Background(), 5, 0)
		if err != nil {
			log.Println(err.Error())
		} else {
			ch <- leader
			log.Println("New TOP 5")
		}
		time.Sleep(10 * time.Second)
	}
}
func updateCacheWorker(cache *LeaderboardCache, ch <-chan []LeaderboardEntry) {
	for newTop := range ch {
		cache.mu.Lock()
		cache.entries = newTop
		cache.mu.Unlock()
	}
}
