package util

import (
	"fmt"
	"time"
)

func ParseTime(input string, formats []string) (time.Time, error) {
	for _, format := range formats {
		t, err := time.Parse(format, input)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized time format %s, supported formats: %s", input, formats)
}
