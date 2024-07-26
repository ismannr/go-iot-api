package models

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
)

type DeviceGrouping struct {
	gorm.Model
	ID             uuid.UUID `gorm:"type:uuid;primary_key"`
	UmkmDataId     uuid.UUID `gorm:"column:umkm_data_id"`
	GroupName      string    `json:"group_name"`
	NumberOfDevice int       `json:"number_of_device"`
}

func CreateGrouping(db *gorm.DB, umkmDataId uuid.UUID, groupName string) (error, string, int) {
	var dg DeviceGrouping

	if err := db.Where("LOWER(group_name) = LOWER(?) AND umkm_data_id = ?", groupName, umkmDataId).First(&DeviceGrouping{}).Error; err == nil {
		return nil, "Group already exist", http.StatusBadRequest
	} else {
		dg.ID = uuid.New()
		dg.UmkmDataId = umkmDataId
		dg.GroupName = groupName
		dg.NumberOfDevice = 0
		if err := db.Create(&dg).Error; err != nil {
			return err, err.Error(), http.StatusInternalServerError
		}
		return nil, "Succesfully creating group", 200
	}
}

func RenameGrouping(db *gorm.DB, umkmDataId uuid.UUID, groupId uuid.UUID, newGroupName string) (error, string, int) {
	var dg DeviceGrouping

	if err := db.Where("id = ? AND umkm_data_id = ?", groupId, umkmDataId).First(&dg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "Group not found", http.StatusNotFound
		}
		return err, "Failed to retrieve group", http.StatusInternalServerError
	}

	if err := db.Model(&dg).Update("GroupName", newGroupName).Error; err != nil {
		return err, "Failed to update group", http.StatusInternalServerError
	}

	if err := db.Model(&Device{}).Where("group_id = ? AND umkm_data_id = ?", groupId, umkmDataId).Update("GroupName", newGroupName).Error; err != nil {
		return err, "Failed to update device group names", http.StatusInternalServerError
	}

	return nil, "Successfully renamed group", http.StatusOK
}

func UnassignDevicesFromGroup(db *gorm.DB, groupID uuid.UUID, userID uuid.UUID) (error, string, int) {
	var message string
	updateFields := map[string]interface{}{
		"group_name": nil,
		"group_id":   nil,
	}

	err := db.Model(&Device{}).
		Where("group_id = ? AND umkm_data_id = ?", groupID, userID).
		Updates(updateFields).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		message = fmt.Sprintf("Group with ID: %s not found", groupID)
		return err, message, http.StatusBadRequest
	} else if err != nil {
		log.Println(err.Error())
		return err, "Failed unassign device from group", http.StatusBadRequest
	}

	return nil, "Successfully unassigned all device from group", http.StatusOK
}
