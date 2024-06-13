package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type BinusianData struct {
	gorm.Model
	ID           uuid.UUID `gorm:"type:uuid;primary_key"`
	BinusianID   string    `json:"binusian_id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Dob          time.Time
	Gender       string
	SystemDataID uint       `gorm:"column:system_data_id"`
	SystemData   SystemData `gorm:"foreignKey:SystemDataID"`
}
