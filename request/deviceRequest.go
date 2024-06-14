package request

import (
	"github.com/google/uuid"
)

type DeviceRequest struct {
	ID          uuid.UUID
	Name        string
	IsActivated bool
	UmkmDataId  uuid.UUID
}
