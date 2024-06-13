package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Token struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key"`
	gorm.Model
	Token        string    `json:"token"`
	TokenExpired bool      `json:"token_expired"`
	TokenExpiry  time.Time `json:"last_login"`
}
