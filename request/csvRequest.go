package request

import (
	"github.com/google/uuid"
	"time"
)

type CSVData struct {
	OxygenLevel float32       `json:"oxygen_level,omitempty"`
	WaterTemp   float32       `json:"water_temp,omitempty"`
	EcLevel     float32       `json:"ec_level,omitempty"`
	PhLevel     float32       `json:"ph_level,omitempty"`
	TimeStamp   time.Time     `json:"time_stamp"`
	ID          uuid.UUID     `json:"id,omitempty"`
	Interval    time.Duration `json:"interval"`
}
