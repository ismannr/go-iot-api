package request

import (
	"time"
)

type CSVData struct {
	OxygenLevel float32       `uri:"oxygen_level"`
	WaterTemp   float32       `uri:"water_temp"`
	EcLevel     float32       `uri:"ec_level"`
	PhLevel     float32       `uri:"ph_level"`
	TimeStamp   time.Time     `uri:"time_stamp" binding:"required"`
	ID          string        `uri:"id" binding:"required,uuid"`
	Interval    time.Duration `json:"interval"`
}
