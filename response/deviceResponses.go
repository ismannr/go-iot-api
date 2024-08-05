package response

import (
	"gin-crud/models"
	"github.com/google/uuid"
)

type DeviceResponse struct {
	ID          uuid.UUID  `json:"id,omitempty"`
	Name        string     `json:"name,omitempty"`
	IsActivated bool       `json:"is_activated"`
	GroupName   *string    `json:"group_name"`
	GroupID     *uuid.UUID `json:"group_id"`
	UmkmDataId  *uuid.UUID `json:"umkm_data_id,omitempty"`
}

func BindDeviceToResponse(device *models.Device) DeviceResponse {
	resp := DeviceResponse{
		ID:          device.ID,
		Name:        device.Name,
		IsActivated: device.IsActivated,
		UmkmDataId:  device.UmkmDataId,
	}
	return resp
}

type DeviceGroupResponse struct {
	ID         uuid.UUID `json:"id"`
	UmkmDataId uuid.UUID `json:"umkm-data-id"`
	GroupName  string    `json:"group-name"`
}
