package models

import "gorm.io/gorm"

type UserPermissions struct {
	gorm.Model
	Email                string `gorm:"foreignKey:Email"`
	UserDeactivated      bool
	BannedFromCommenting bool
	BannedFromPosting    bool
	BannedFromAnalytix   bool
}
