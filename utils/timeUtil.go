package utils

import (
	"fmt"
	"time"
)

// ParseDate parses a string in the format "YYYY-MM-DD" to a time.Time object.
func ParseDate(dateString string) (time.Time, error) {
	layout := "2006-01-02"
	date, err := time.Parse(layout, dateString)
	if err != nil {
		return time.Time{}, err
	}
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)

	location, err := time.LoadLocation("Asia/Bangkok") // Use appropriate location string
	if err != nil {
		fmt.Println("Error loading location:", err)
		return time.Time{}, err
	}
	return date.In(location), nil
}

func IsAdult(dob string) bool {
	strDob, err := ParseDate(dob)
	if err != nil {
		return false
	}
	minimumAdultAge := 17
	return time.Now().Sub(strDob) >= time.Duration(minimumAdultAge)*365*24*time.Hour
}

func CurrentTimeWIB() time.Time {
	location := time.FixedZone("GMT+7", 7*60*60) // 7 hours ahead of UTC
	currentTime := time.Now().In(location)
	return currentTime
}
