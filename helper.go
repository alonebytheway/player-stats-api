package main

import (
	"log"
	"time"
)

func CalculateKD(kills int, deaths int) float64 {
	if deaths == 0 {
		return float64(kills)
	} else {
		return float64(kills) / float64(deaths)
	}
}

func backgroundWorker() {
	for {
		log.Println("Background task: recalculating the global ranking...")
		time.Sleep(10 * time.Second)
	}
}
