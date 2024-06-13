package request

import (
	"github.com/google/uuid"
)

type DeviceRequest struct {
	ID          uuid.UUID
	Password    string
	OxygenLevel float32
	WaterTemp   float32
	EcLevel     float32
	PhLevel     float32
	IsActivated bool
	UmkmDataId  uuid.UUID
}
