package db

import (
	"fmt"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB, models ...any) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	return nil
}
