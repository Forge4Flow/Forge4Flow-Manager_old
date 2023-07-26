package migrate

import (
	"log"

	"github.com/forge4flow/forge4flow-manager/internal/models"

	"gorm.io/gorm"
)

func Migrate(instance *gorm.DB) {
	instance.AutoMigrate(
		&models.CoreInstance{},
	)

	log.Println("Database Migration Completed!")
}
