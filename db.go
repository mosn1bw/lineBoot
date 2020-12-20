package main

import (
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

var repo *gorm.DB

func NewDB() *gorm.DB {
	dbConfig := os.Getenv("DATABASE_URL")
	db, err := gorm.Open("postgres", dbConfig)

	if err != nil {
		log.Fatal(err)
	}

	return db
}
