package request

import (
	"github.com/google/uuid"
)

type DeviceRequest struct {
	ID          uuid.UUID
	Name        string
	GroupID     uuid.UUID `json:"group_id"`
	IsActivated bool
	UmkmDataId  uuid.UUID
}
