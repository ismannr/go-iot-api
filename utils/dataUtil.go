package utils

import (
	"crypto/rand"
	"encoding/hex"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func IsPDF(filename string) bool {
	extension := filepath.Ext(filename)
	return strings.ToLower(extension) == ".pdf"
}

func IsImage(filename string) bool {
	extension := filepath.Ext(filename)
	return strings.ToLower(extension) == ".jpg" || strings.ToLower(extension) == ".jpeg" || strings.ToLower(extension) == ".png"
}

func GenerateUniqueFileName(fileExt string) string {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		// Handle error
		return ""
	}

	randomString := hex.EncodeToString(randomBytes)
	timestamp := time.Now().UnixNano()
	fileName := randomString + "_" + strconv.FormatInt(timestamp, 10) + fileExt

	return fileName
}
