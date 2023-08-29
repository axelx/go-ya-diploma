package utils

import (
	"fmt"
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
	floatType := reflect.TypeOf(float64(0))
	v := reflect.ValueOf(unk)
	v = reflect.Indirect(v)
	if !v.Type().ConvertibleTo(floatType) {
		fmt.Errorf("cannot convert %v to float64", v.Type())
	}
	fv := v.Convert(floatType)
	return fv.Float()
}
