package request

type MonitoringRequest struct {
	Date     string `json:"date" binding:"required"`
	DeviceID string `json:"device_id"`
	Interval string `json:"interval"`
}
