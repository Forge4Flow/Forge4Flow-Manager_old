package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Connect(connectionString string) (*gorm.DB, error) {
	instance, err := gorm.Open(sqlite.Open("Forge4FlowManager.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	log.Println("Connected to Database!")

	return instance, nil
}
