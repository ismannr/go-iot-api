package utils

import (
	"time"
)

// ParseDate parses a string in the format "YYYY-MM-DD" to a time.Time object.
func ParseDate(dateString string) (time.Time, error) {
	layout := "2006-01-02"
	parsedTime, err := time.Parse(layout, dateString)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
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
