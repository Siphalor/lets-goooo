package util

import (
	"fmt"
	"time"
)

// GetDateFilename creates a sortable file name that is unique for each day.
func GetDateFilename(time time.Time) string {
	year, month, day := time.Date()
	return fmt.Sprintf("%04d%02d%02d", year, month, day)
}
