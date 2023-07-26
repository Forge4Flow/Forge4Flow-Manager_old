package database

import (
	"log"

	"github.com/forge4flow/forge4flow-manager/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() error {
	instance, err := gorm.Open(sqlite.Open("Forge4FlowManager.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	DB = instance

	log.Println("Connected to Database!")

	DB.AutoMigrate(
		&models.CoreInstance{},
	)

	log.Println("Database Migration Completed!")

	return nil
}
