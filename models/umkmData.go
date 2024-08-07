package models

import (
	"errors"
	"gin-crud/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

type UmkmData struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key"`
	gorm.Model
	Name           string      `json:"name"`
	Email          string      `json:"email" gorm:"uniqueIndex"`
	Gender         string      `json:"gender"`
	Phone          string      `json:"phone" gorm:"uniqueIndex"`
	Dob            time.Time   `json:"birth_date"`
	Address        string      `json:"address"`
	City           string      `json:"city"`
	Province       string      `json:"province"`
	BusinessName   string      `json:"business_name"`
	BusinessDesc   string      `json:"business_desc"`
	SystemDataID   *uuid.UUID  `gorm:"column:system_data_id;uniqueIndex"`
	SystemData     *SystemData `gorm:"foreignKey:SystemDataID;constraint:OnDelete:CASCADE;"`
	Devices        *[]Device
	DeviceGrouping *[]DeviceGrouping
}

func DeleteDeviceById(db *gorm.DB, userID uuid.UUID, deviceID uuid.UUID) error {
	var device Device
	if err := db.First(&device, "id = ?", deviceID).Error; err != nil {
		return err
	}

	var user UmkmData
	if err := db.Preload("Devices").First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	var deviceFound bool
	for _, dev := range *user.Devices {
		if dev.ID == deviceID {
			deviceFound = true
			break
		}
	}
	if !deviceFound {
		return utils.ErrDeviceAlreadyDeleted
	}

	var updatedDevices []Device
	for _, dev := range *user.Devices {
		if dev.ID != deviceID {
			updatedDevices = append(updatedDevices, dev)
		}
	}
	user.Devices = &updatedDevices

	dataString := string(device.Data)
	lines := strings.Split(dataString, "\n")
	if len(lines) > 0 {
		device.Data = []byte(lines[0] + "\n")
	} else {
		device.Data = nil
	}
	device.IsActivated = false
	device.UmkmDataId = nil
	device.Name = ""

	if err := db.Save(&device).Error; err != nil {
		return err
	}
	return db.Save(&user).Error
}

func RegisterDeviceById(db *gorm.DB, userID uuid.UUID, deviceID uuid.UUID, deviceName string, groupId uuid.UUID) (error, string, int) {
	var device Device
	if err := db.First(&device, "id = ? AND umkm_data_id IS NULL", deviceID).Error; err != nil {
		return err, "Cannot find device", http.StatusBadRequest
	}

	var user *UmkmData
	if err := db.Preload("Devices").First(&user, "id = ?", userID).Error; err != nil {
		return err, "Cannot find user", http.StatusInternalServerError
	}

	for _, dev := range *user.Devices {
		if dev.ID == deviceID {
			return utils.ErrDeviceAlreadyRegistered, "Device already registered", http.StatusBadRequest
		}
	}
	device.Name = deviceName
	device.IsActivated = true
	device.UmkmDataId = &user.ID
	if err := db.Save(&device).Error; err != nil {
		return err, "Failed to save device", http.StatusInternalServerError
	}

	if user.Devices == nil {
		user.Devices = &[]Device{}
	}
	*user.Devices = append(*user.Devices, device)

	err, message, status := AssignDeviceToGroup(db, deviceID, userID, groupId)
	if err != nil {
		return err, message, status
	}
	return db.Save(&user).Error, "Successfully adding device", http.StatusOK
}

func UpdateDeviceName(db *gorm.DB, userID uuid.UUID, deviceID uuid.UUID, newName string) error {
	var device Device
	if err := db.Where("id = ? AND umkm_data_id = ?", deviceID, userID).First(&device).Error; err != nil {
		return gorm.ErrRecordNotFound
	}

	device.Name = newName

	return db.Save(&device).Error
}

func GetUserDeviceById(db *gorm.DB, userID uuid.UUID, deviceID uuid.UUID) (*Device, error) {
	var device Device
	if err := db.Where("id = ? AND umkm_data_id = ?", deviceID, userID).First(&device).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &device, nil
}

func GetAllUserDevices(db *gorm.DB, userID uuid.UUID) ([]Device, error) {
	var devices []Device
	err := db.Where("umkm_data_id = ?", userID).Find(&devices).Error
	if err != nil {
		return nil, err
	}
	return devices, nil
}
