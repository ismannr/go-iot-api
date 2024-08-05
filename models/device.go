package models

import (
	"errors"
	"fmt"
	"gin-crud/initializers"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
)

type Device struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primary_key"`
	Name        string
	Data        []byte
	IsActivated bool
	GroupName   *string    `json:"group_name"`
	GroupID     *uuid.UUID `gorm:"column:group_id"`
	UmkmDataId  *uuid.UUID `gorm:"column:umkm_data_id"`
}

func AssignDeviceToGroup(db *gorm.DB, deviceID uuid.UUID, userID uuid.UUID, groupId uuid.UUID) (error, string, int) {
	var device Device
	var deviceGrouping *DeviceGrouping
	var message string

	if err := db.Where("id = ? AND umkm_data_id = ?", deviceID, userID).First(&device).Error; err != nil {
		message = fmt.Sprintf("Device with id %s not found", deviceID)
		return err, message, http.StatusBadRequest
	}

	if device.GroupName != nil {
		message = fmt.Sprintf("Device already assigned to %s", *device.GroupName)
		return nil, message, http.StatusBadRequest
	}

	if err := db.Where("id = ?", groupId).First(&deviceGrouping).Error; err != nil {
		message = fmt.Sprintf("Group with name %s not found", groupId)
		return err, message, http.StatusBadRequest
	}

	device.GroupName = &deviceGrouping.GroupName
	device.GroupID = &deviceGrouping.ID
	deviceGrouping.NumberOfDevice += 1
	if err := db.Save(&device).Error; err != nil {
		message = fmt.Sprintf("Failed to group the device to %s", groupId)
		return err, message, http.StatusInternalServerError
	}

	if err := db.Save(&deviceGrouping).Error; err != nil {
		message = fmt.Sprintf("Failed to add device count to %s", groupId)
		return err, message, http.StatusInternalServerError
	}

	message = fmt.Sprintf("Succesfully adding device to %s", deviceGrouping.GroupName)
	return nil, message, 200
}

func UnassignDeviceFromGroup(db *gorm.DB, deviceID uuid.UUID, userID uuid.UUID) (error, string, int) {
	var device Device
	var message string
	var deviceGrouping DeviceGrouping

	if err := db.Where("id = ? AND umkm_data_id = ?", deviceID, userID).First(&device).Error; err != nil {
		message = fmt.Sprintf("Device with id %s not found", deviceID)
		return err, message, http.StatusBadRequest
	}

	if device.GroupName == nil && device.GroupID == nil {
		message = "Device is not assigned to any group"
		return nil, message, http.StatusBadRequest
	}

	if err := db.Where("id = ?", device.GroupID).First(&deviceGrouping).Error; err != nil {
		log.Println(err.Error())
		return err, "Failed to get device group", http.StatusInternalServerError
	}

	deviceGrouping.NumberOfDevice -= 1

	device.GroupName = nil
	device.GroupID = nil

	if err := db.Save(&deviceGrouping).Error; err != nil {
		return err, "Failed to update device count", http.StatusInternalServerError
	}

	if err := db.Save(&device).Error; err != nil {
		message = "Failed to unassign the device from the group"
		return err, message, http.StatusInternalServerError
	}

	message = "Successfully unassigned the device from the group"
	return nil, message, http.StatusOK
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
