package util

import (
	"strconv"
	"time"
)

func ParseTime(t string) (time.Time, error) {
	// time contain day and time
	parsedTime, err := time.Parse("2006-01-02 15:04:05.000000", t)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}

func ParseToInt(s string) (int, error) {
	// convert string to int
	parsedInt, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return parsedInt, nil
}
