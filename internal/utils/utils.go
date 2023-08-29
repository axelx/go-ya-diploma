package utils

import (
	"reflect"
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

func GetFloat(unk interface{}) float64 {
	if unk == nil {
		return 0.0
	}
	floatType := reflect.TypeOf(float64(0))
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	fv := v.Convert(floatType)
	return fv.Float()
}
