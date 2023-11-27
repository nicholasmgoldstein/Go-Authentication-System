package initializers

import (
	"github.com/auth/models"
)

func SyncDB() {
	DB.AutoMigrate(
		&models.User{},
		&models.UserPermissions{},
	)
}
