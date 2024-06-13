package models

import (
	"errors"
	"gin-crud/initializers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Device struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	Data        []byte
	IsActivated bool
	UmkmDataId  *uuid.UUID `gorm:"column:umkm_data_id"`
}

func SaveCSVToDevice(db *gorm.DB, csvData []byte, deviceID uuid.UUID) error {
	var device Device

	if err := db.First(&device, "id = ?", deviceID).Error; err != nil {
		return err
	}
	device.Data = append(device.Data, csvData...)

	if err := db.Save(&device).Error; err != nil {
		return err
	}

	return nil
}

func GetDeviceById(db *gorm.DB, deviceID uuid.UUID) (*Device, error) {
	var device Device
	if err := db.Where("id = ?", deviceID).First(&device).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &device, nil
}

func (device *Device) BeforeCreate(tx *gorm.DB) (err error) {
	headers := []byte("oxygen_level,water_temp,ec_level,ph_level,time_stamp,id\n")
	device.Data = headers
	return
}

func CreateDevice(device *Device) error {
	if err := initializers.DB.Create(device).Error; err != nil {
		return err
	}
	return nil
}
