package models

import "time"

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
	Order    []Order
}

type Order struct {
	ID         int        `json:"id,omitempty"`
	Number     string     `json:"number,omitempty"`
	Accrual    float64    `json:"accrual"`
	Withdrawn  float64    `json:"sum"`
	Status     string     `json:"status,omitempty"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`
	UserID     int        `json:"user_id,omitempty"`
}
type OrderWithdrawal struct {
	ID         int        `json:"id,omitempty"`
	Number     string     `json:"order,omitempty"`
	Withdrawn  float64    `json:"sum"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}
