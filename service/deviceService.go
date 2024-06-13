package service

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"gin-crud/initializers"
	"gin-crud/models"
	"gin-crud/request"
	"gin-crud/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func toCSV(data request.CSVData) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	record := []string{
		strconv.FormatFloat(float64(data.OxygenLevel), 'f', -1, 32),
		strconv.FormatFloat(float64(data.WaterTemp), 'f', -1, 32),
		strconv.FormatFloat(float64(data.EcLevel), 'f', -1, 32),
		strconv.FormatFloat(float64(data.PhLevel), 'f', -1, 32),
		data.TimeStamp.Format(time.RFC3339),
		data.ID.String(),
	}

	if err := writer.Write(record); err != nil {
		return nil, err
	}

	writer.Flush()
	return buf.Bytes(), nil
}

func ReceiveAndSaveData(c *gin.Context) {
	var csvData request.CSVData

	if err := c.Bind(&csvData); err != nil {
		log.Println("Invalid JSON data")
		response.GlobalResponse(c, "Invalid JSON data", http.StatusBadRequest, nil)
		return
	}

	device, err := models.GetDeviceById(initializers.DB, csvData.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			message := fmt.Sprintf("Device not found ID:%s", csvData.ID.String())
			log.Println(message)
			response.GlobalResponse(c, message, http.StatusNotFound, nil)
		} else {
			message := fmt.Sprintf("Failed to retrieve device ID:%s", csvData.ID.String())
			log.Println(message)
			response.GlobalResponse(c, message, http.StatusInternalServerError, nil)
		}
		return
	}

	if device.UmkmDataId == nil {
		message := fmt.Sprintf("Device not associated with any user ID:%s", csvData.ID.String())
		log.Println(message)
		response.GlobalResponse(c, message, http.StatusBadRequest, nil)
		return
	}

	csvBytes, err := toCSV(csvData)
	if err != nil {
		message := fmt.Sprintf("Failed to convert data to CSV ID:%s", csvData.ID.String())
		log.Println(message)
		response.GlobalResponse(c, message, http.StatusInternalServerError, nil)
		return
	}

	err = models.SaveCSVToDevice(initializers.DB, csvBytes, csvData.ID)
	if err != nil {
		message := fmt.Sprintf("Failed to save data to database ID:%s", csvData.ID.String())
		log.Println(message)
		response.GlobalResponse(c, message, http.StatusInternalServerError, nil)
		return
	}

	message := "Successfully appended data to CSV"
	response.GlobalResponse(c, message, http.StatusOK, nil)
}

func GetMonitoringData(c *gin.Context) {
	var req request.MonitoringRequest
	var targetDate time.Time
	var interval time.Duration
	user, err := getUmkmByAuth(c)
	if err != nil {
		response.GlobalResponse(c, "Invalid user", http.StatusUnauthorized, nil)
		return
	}

	if err := c.Bind(&req); err != nil {
		response.GlobalResponse(c, "Invalid date data", http.StatusBadRequest, nil)
		return
	}
	deviceID, err := uuid.Parse(req.DeviceID)
	if err != nil {
		response.GlobalResponse(c, "Invalid device ID format", http.StatusBadRequest, nil)
		return
	}

	if req.Date == "" || req.Interval == "" {
		targetDate = time.Now()
		interval = time.Minute * 5

	} else {
		targetDate, err = time.Parse(time.RFC3339, req.Date)
		if err != nil {
			response.GlobalResponse(c, "Invalid date format. Use yyyy-mm-dd", http.StatusBadRequest, nil)
			return
		}
		if targetDate.After(time.Now().AddDate(0, 0, 1)) {
			response.GlobalResponse(c, "Date must not be after today", http.StatusBadRequest, nil)
			return
		}

		interval, err = time.ParseDuration(req.Interval)
		if err != nil {
			response.GlobalResponse(c, "Invalid time format. Use 0h0m0s", http.StatusBadRequest, nil)
			return
		}

		if interval > time.Minute*30 || interval < time.Minute*0 {
			response.GlobalResponse(c, "Cannot greater than 30 minutes and lower than 0", http.StatusBadRequest, nil)
			return
		}
	}

	csvData, err := GetDeviceCsvData(user.ID, deviceID, targetDate, interval)
	if err != nil {
		log.Println(err.Error())
		response.GlobalResponse(c, "Failed to retrieve CSV data", http.StatusInternalServerError, nil)
		return
	}

	response.GlobalResponse(c, "Successfully retrieved CSV data", http.StatusOK, csvData)
}

func parseCSVData(data []byte) ([]request.CSVData, error) {
	var records []request.CSVData
	reader := csv.NewReader(bytes.NewReader(data))

	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		oxygenLevel, err := strconv.ParseFloat(record[0], 32)
		if err != nil {
			return nil, err
		}

		waterTemp, err := strconv.ParseFloat(record[1], 32)
		if err != nil {
			return nil, err
		}

		ecLevel, err := strconv.ParseFloat(record[2], 32)
		if err != nil {
			return nil, err
		}

		phLevel, err := strconv.ParseFloat(record[3], 32)
		if err != nil {
			return nil, err
		}

		timeStamp, err := time.Parse(time.RFC3339, record[4])
		if err != nil {
			return nil, err
		}

		id, err := uuid.Parse(record[5])
		if err != nil {
			return nil, err
		}

		records = append(records, request.CSVData{
			OxygenLevel: float32(oxygenLevel),
			WaterTemp:   float32(waterTemp),
			EcLevel:     float32(ecLevel),
			PhLevel:     float32(phLevel),
			TimeStamp:   timeStamp,
			ID:          id,
		})
	}

	return records, nil
}

func filterCSVData(records []request.CSVData, targetDate time.Time, interval time.Duration) []request.CSVData {
	var filteredRecords []request.CSVData

	for _, record := range records {
		if record.TimeStamp.After(targetDate.Add(-interval)) && record.TimeStamp.Before(targetDate.Add(1*time.Second)) {
			filteredRecords = append(filteredRecords, record)
		}
	}

	return filteredRecords
}

func writeFilteredCSVData(records []request.CSVData) (string, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	headers := []string{"OxygenLevel", "WaterTemp", "EcLevel", "PhLevel", "TimeStamp", "ID"}
	if err := w.Write(headers); err != nil {
		return "", err
	}

	for _, record := range records {
		row := []string{
			strconv.FormatFloat(float64(record.OxygenLevel), 'f', -1, 32),
			strconv.FormatFloat(float64(record.WaterTemp), 'f', -1, 32),
			strconv.FormatFloat(float64(record.EcLevel), 'f', -1, 32),
			strconv.FormatFloat(float64(record.PhLevel), 'f', -1, 32),
			record.TimeStamp.Format(time.RFC3339),
			record.ID.String(),
		}
		if err := w.Write(row); err != nil {
			return "", err
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func GetDeviceCsvData(userID, deviceID uuid.UUID, targetDate time.Time, interval time.Duration) (string, error) {
	var device models.Device

	if err := initializers.DB.Where("id = ? AND umkm_data_id = ?", deviceID, userID).First(&device).Error; err != nil {
		return "", err
	}

	records, err := parseCSVData(device.Data)
	if err != nil {
		return "", err
	}

	filteredRecords := filterCSVData(records, targetDate, interval)
	log.Println(targetDate)
	filteredCSV, err := writeFilteredCSVData(filteredRecords)
	if err != nil {
		return "", err
	}

	return filteredCSV, nil
}
