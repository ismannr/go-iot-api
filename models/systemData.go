package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type SystemData struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key"`
	gorm.Model
	Email           string                 `json:"email"`
	Password        string                 `json:"-" json:"password"`
	Role            Role                   `json:"role"`
	Level           Level                  `json:"level"`
	CurrentlyLogin  bool                   `json:"currently_login"`
	RecoveryTokenId *uuid.UUID             `gorm:"column:recovery_token_id;uniqueIndex"`
	RecoveryToken   *PasswordRecoveryToken `gorm:"foreignKey:RecoveryTokenId;constraint:OnDelete:SET NULL;"`
	TokenID         *uuid.UUID             `gorm:"column:token;uniqueIndex"`
	Token           *Token                 `gorm:"foreignKey:TokenID;constraint:OnDelete:SET NULL;"`
	LastLogin       time.Time              `json:"last_login"`
}

func (u *SystemData) BeforeDelete(tx *gorm.DB) (err error) {
	if u.RecoveryToken != nil {
		if err := tx.Unscoped().Delete(u.RecoveryToken).Error; err != nil {
			return err
		}
	}

	if u.Token != nil {
		if err := tx.Unscoped().Delete(u.Token).Error; err != nil {
			return err
		}
	}

	return nil
}
