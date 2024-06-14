package response

import (
	"gin-crud/models"
	"github.com/google/uuid"
)

type DeviceResponse struct {
	ID          uuid.UUID  `json:"id,omitempty"`
	Name        string     `json:"name,omitempty"`
	IsActivated bool       `json:"is_activated,omitempty"`
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
