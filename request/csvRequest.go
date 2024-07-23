package request

import (
	"time"
)

type CSVData struct {
	OxygenLevel float32       `uri:"oxygen_level" binding:"required"`
	WaterTemp   float32       `uri:"water_temp" binding:"required"`
	EcLevel     float32       `uri:"ec_level" binding:"required"`
	PhLevel     float32       `uri:"ph_level" binding:"required"`
	TimeStamp   time.Time     `uri:"time_stamp" binding:"required"`
	ID          string        `uri:"id" binding:"required,uuid"`
	Interval    time.Duration `json:"interval"`
}
