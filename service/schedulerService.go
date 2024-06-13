package service

import (
	"errors"
	"fmt"
	"gin-crud/initializers"
	model "gin-crud/models"
	"gorm.io/gorm"
	"log"
	"time"
)

func TokenExpirationCheckAndUpdateScheduler() {
	tokenExpirationCheckAndUpdate()
	ticker := time.NewTicker(time.Minute * 30)
	defer ticker.Stop()
	for range ticker.C {
		tokenExpirationCheckAndUpdate()
	}
}

func tokenExpirationCheckAndUpdate() {
	var sysData []model.SystemData

	initializers.DB.Where("last_login < ?", time.Now().Add(-1*time.Hour)).Find(&sysData)

	for _, session := range sysData {
		session.CurrentlyLogin = false

		if err := initializers.DB.Save(&session).Error; err != nil {
			log.Println("Failed to update session:", err)
			continue
		}

		if session.TokenID == nil {
			continue
		}

		var token model.Token
		if err := initializers.DB.First(&token, session.TokenID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			log.Println("Failed to retrieve token data:", err)
			continue
		}

		if err := initializers.DB.Unscoped().Delete(&token).Error; err != nil {
			log.Println("Failed to invalidate token:", err)
			continue
		}
	}
}

func ClearDeviceDataScheduler() {
	now := time.Now().UTC()

	nextFirstOfMonth := time.Date(now.Year(), now.Month(), 1, 23, 59, 0, 0, time.UTC)
	if now.Day() == 1 {
		nextFirstOfMonth = nextFirstOfMonth.AddDate(0, 1, 0) // Move to next month if today is the first day
	}
	durationUntilNext := nextFirstOfMonth.Sub(now)

	time.Sleep(durationUntilNext)

	ticker := time.Tick(24 * time.Hour)

	for range ticker {
		if time.Now().UTC().Day() == 1 && time.Now().UTC().Hour() == 23 && time.Now().UTC().Minute() == 59 {
			message := fmt.Sprintln("Running scheduler to clear Device Data...")
			log.Println(message)
			err := clearDeviceData()
			if err != nil {
				message = fmt.Sprintf("Error clearing Device Data: %v\n", err)
				log.Println(message)
			} else {
				message = fmt.Sprintln("Device Data cleared successfully.")
				log.Println(message)
			}

			nextFirstOfMonth = time.Now().UTC().AddDate(0, 1, 0)
			nextFirstOfMonth = time.Date(nextFirstOfMonth.Year(), nextFirstOfMonth.Month(), 1, 23, 59, 0, 0, time.UTC)

			durationUntilNext = nextFirstOfMonth.Sub(time.Now().UTC())

			time.Sleep(durationUntilNext)
		}
	}
}

func clearDeviceData() error {
	var devices []model.Device

	if err := initializers.DB.Find(&devices).Error; err != nil {
		return err
	}

	for i := range devices {
		devices[i].Data = nil
		if err := initializers.DB.Save(&devices[i]).Error; err != nil {
			return err
		}
	}

	return nil
}
