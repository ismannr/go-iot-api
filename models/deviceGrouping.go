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
	ID         uuid.UUID `gorm:"type:uuid;primary_key"`
	UmkmDataId uuid.UUID `gorm:"column:umkm_data_id"`
	GroupName  string    `json:"group_name"`
}

func CreateGrouping(db *gorm.DB, umkmDataId uuid.UUID, groupName string) (error, string, int) {
	var dg DeviceGrouping

	if err := db.Where("group_name = ? AND umkm_data_id = ?", groupName, umkmDataId).First(&DeviceGrouping{}).Error; err == nil {
		return nil, "Group already exist", http.StatusBadRequest
	} else {
		dg.ID = uuid.New()
		dg.UmkmDataId = umkmDataId
		dg.GroupName = groupName

		if err := db.Create(&dg).Error; err != nil {
			return err, err.Error(), http.StatusInternalServerError
		}
		return nil, "Succesfully creating group", 200
	}
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

	return nil, "Successfully unassigned devices from group", http.StatusOK
}
