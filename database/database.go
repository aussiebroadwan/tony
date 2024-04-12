package database

import (
	"fmt"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	log "github.com/sirupsen/logrus"
)

func NewDatabase() *gorm.DB {
	host := os.Getenv("DB_HOST")
	dbname := os.Getenv("DB_NAME")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	if host == "" {
		log.Fatal("No host provided. Please set DB_HOST environment variable.")
	}

	if dbname == "" {
		log.Fatal("No database name provided. Please set DB_NAME environment variable.")
	}

	if user == "" {
		log.Fatal("No user provided. Please set DB_USER environment variable.")
	}

	if password == "" {
		log.Fatal("No password provided. Please set DB_PASSWORD environment variable.")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", host, user, password, dbname)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return db
}
