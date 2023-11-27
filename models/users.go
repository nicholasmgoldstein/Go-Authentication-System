package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"unique"`
	Name     string
	Password string
	Pic      string
	Intro    string
	RefRank  uint
	DOB      time.Time
	Country  string
	Location string
}
