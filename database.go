package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func ConnectDB() *sql.DB {
	connStr := "host=db port=5432 user=postgres password=admin123 dbname=postgres sslmode=disable"

	var db *sql.DB
	var err error

	for i := 0; i < 10; i++ {
		// Используем "=", а не ":=", чтобы не создавать новые переменные
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
		}

		if err == nil {
			log.Println("Success: connected to DB")
			return db // Возвращаем объект базы
		}

		log.Printf("DB is not ready (attempt %d/10)... waiting 2s", i+1)
		time.Sleep(2 * time.Second)
	}

	panic("Could not connect to DB after 10 attempts: " + err.Error())
}
