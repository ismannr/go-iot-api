package utils

import "errors"

var (
	ErrDeviceAlreadyRegistered = errors.New("device is already registered for the user")
	ErrDeviceAlreadyDeleted    = errors.New("device not found")
	ErrDeviceNotFound          = errors.New("device not found")
)
