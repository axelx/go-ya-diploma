package models

import "time"

type MetricType string

var MetricGauge MetricType = "gauge"
var MetricCounter MetricType = "counter"

type Metrics struct {
	ID    string     `json:"id"`              // имя метрики
	MType MetricType `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64     `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64   `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type User struct {
	ID       string `json:"id"`
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
	Order    []Order
}

type Order struct {
	ID         int        `json:"id,omitempty"`
	Number     string     `json:"number,omitempty"`
	Accrual    int        `json:"accrual"`
	Withdrawn  int        `json:"withdrawn"`
	Status     string     `json:"status,omitempty"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`
	UserID     int        `json:"user_id,omitempty"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
