package utils

import (
	"strconv"
	"time"
)

func StrToInt(s string) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return 0
}
func IntToStr(i int) string {
	s := strconv.Itoa(i)
	return s
}

func TimeFormat(s *time.Time) string {
	return s.Format(time.RFC3339)
}
