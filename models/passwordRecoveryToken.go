package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasswordRecoveryToken struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key"`
	gorm.Model
	RandTokenAccess string `json:"token-access"`
}
