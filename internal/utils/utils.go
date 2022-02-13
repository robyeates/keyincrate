package utils

import (
	"strconv"
	"time"
)

func EpochStringToTime(s string) (time.Time, error) {
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(sec, 0), nil
}

func EpochIntToTime(i int) (time.Time, error) {
	return time.Unix(int64(i), 0), nil
}
