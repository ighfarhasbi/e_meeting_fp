package utils

import (
	"strconv"
	"time"
)

func StringToInt(s1, s2, s3 string) (int, int, int, error) {
	// looping konversi string ke int
	var int1, int2, int3 int
	var err error
	for i := range 4 {
		switch i {
		case 0:
			int1, err = strconv.Atoi(s1)
		case 1:
			int2, err = strconv.Atoi(s2)
		case 2:
			int3, err = strconv.Atoi(s3)
		}
		if err != nil {
			return 0, 0, 0, err
		}
	}
	return int1, int2, int3, nil

}

func StringToTimestamptz(input string) (time.Time, error) {
	layout := "2006-01-02 15:04:05.000 -0700"
	t, err := time.Parse(layout, input)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func CalculateDuration(start, end time.Time) time.Duration {
	return end.Sub(start)
}

func IsOverlapping(start1, end1 time.Time, startTimesDB, endTimesDB []time.Time) bool {
	for i := range startTimesDB {
		start2 := startTimesDB[i]
		end2 := endTimesDB[i]

		// Dua interval overlap jika start1 < end2 DAN start2 < end1
		if start1.Before(end2) && start2.Before(end1) {
			return true
		}
	}
	return false
}
