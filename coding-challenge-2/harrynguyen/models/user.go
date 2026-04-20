package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents the user model in the database
type User struct {
	gorm.Model
	Username     string     `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"not null" json:"-"`
	PhoneNumber  string     `gorm:"uniqueIndex;not null" json:"phone_number"`
	Status       string     `gorm:"default:'pending'" json:"status"`
	OTPCode      string     `json:"-"`
	OTPExpiresAt *time.Time `json:"-"`
	Country      string     `json:"country"`
}

const (
	UserStatusPending = "pending"
	UserStatusActive  = "active"
)
