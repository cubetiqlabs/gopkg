package util

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cubetiqlabs/gopkg/types"
)

func ParseDuration(input string) (time.Duration, error) {
	unit := strings.TrimLeft(input, "0123456789.")
	valueStr := strings.TrimSuffix(input, unit)
	if valueStr == "" {
		return 0, fmt.Errorf("invalid duration format: %q", input)
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid duration value: %q", input)
	}

	switch strings.ToLower(unit) {
	case "ns":
		return time.Duration(value) * time.Nanosecond, nil
	case "us", "µs":
		return time.Duration(value) * time.Microsecond, nil
	case "ms":
		return time.Duration(value) * time.Millisecond, nil
	case "s":
		return time.Duration(value) * time.Second, nil
	case "m":
		return time.Duration(value) * time.Minute, nil
	case "h":
		return time.Duration(value) * time.Hour, nil
	case "d":
		return time.Duration(value) * time.Hour * 24, nil // Equivalent to 1 day
	case "w":
		return time.Duration(value) * time.Hour * 24 * 7, nil // Equivalent to 1 week
	default:
		return 0, fmt.Errorf("unknown unit: %q", unit)
	}
}

// ParseDateRange parses the start and end date strings into a DateRange struct.
// The date format is expected to be "YYYY-MM-DD".
func ParseDateRange(startDate, endDate string, includeTime bool) (*types.DateRange, error) {
	if includeTime {
		// Parse the date range with time included with start of time as 00:00:00
		startTime, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, fmt.Errorf("start_date: %v", err)
		}

		// Parse the date range with time included with end of time as 23:59:59
		endTime, err := time.Parse("2006-01-02 15:04:05", endDate+" 23:59:59")
		if err != nil {
			return nil, fmt.Errorf("end_date: %v", err)
		}

		return &types.DateRange{
			StartDate: startTime,
			EndDate:   endTime,
		}, nil
	}

	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("start_date: %v", err)
	}

	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("end_date: %v", err)
	}

	return &types.DateRange{
		StartDate: startTime,
		EndDate:   endTime,
	}, nil
}
