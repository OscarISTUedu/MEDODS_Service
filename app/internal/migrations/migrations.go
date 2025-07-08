package migrations

import (
	"github.com/OscarISTUedu/MEDODS_Service/internal/database/models"
	"gorm.io/gorm"
)

func AutoMigrateAll(db *gorm.DB) error {
	return db.AutoMigrate(models.GetRegisteredModels()...)
}
