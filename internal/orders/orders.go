package orders

import (
	"github.com/axelx/go-ya-diploma/internal/core"
	"github.com/axelx/go-ya-diploma/internal/models"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type Order struct {
	ID         int        `json:"id,omitempty"`
	Number     string     `json:"number,omitempty"`
	Accrual    int        `json:"accrual"`
	Withdrawn  int        `json:"withdrawn"`
	Status     string     `json:"status,omitempty"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`
	UserID     int        `json:"user_id,omitempty"`
}

func (o Order) SearchOne(db *sqlx.DB, lg *zap.Logger, orderNum string) (int, string) {

	ord, err := core.FindOrder(db, lg, orderNum)
	if err != nil {
		lg.Error("order SearchOne", zap.String("err", err.Error()))
		return 0, ""
	}
	return ord.UserID, ord.Number
}

func (o Order) SearchMany(s string) ([]int, []string) {
	return []int{5}, []string{"order_" + s}
}

func (o Order) Create(db *sqlx.DB, lg *zap.Logger, login, password string) error {
	err := core.CreateNewUser(db, lg, login, password)
	return err
}

func (o Order) Talk() string {
	return "order talk"
}

func LunaCheck(order string, lg *zap.Logger) bool {
	sum := 0
	for i := len(order); i > 0; i-- {
		num, err := strconv.Atoi(string(order[i-1]))
		if err != nil {
			lg.Info("orders LunaCheck ошибка", zap.String("about", ""))
			return false
		}
		digit := num
		if (i)%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	return sum%10 == 0
}

func FindOrder(db *sqlx.DB, lg *zap.Logger, orderID string) (models.Order, error) {
	o, err := core.FindOrder(db, lg, orderID)
	if err != nil {
		lg.Error("order FindOrder", zap.String("err", err.Error()))
		return o, err
	}
	return o, err
}

func FindOrders(db *sqlx.DB, lg *zap.Logger, userID int) ([]models.Order, error) {
	os, err := core.FindOrders(db, lg, userID)
	if err != nil {
		lg.Error("order FindOrders", zap.String("err", err.Error()))
	}

	return os, nil
}

func AddOrder(db *sqlx.DB, lg *zap.Logger, userID int, orderID string, withdrawn float64, chAdd chan string) error {
	err := core.AddOrder(db, lg, userID, orderID, withdrawn)
	if err == nil {
		lg.Info("order AddOrder and add to channel", zap.String("about", ""))
		chAdd <- orderID
	}
	return err
}

func UpdateStatus(db *sqlx.DB, lg *zap.Logger, orderID, status string) error {
	err := core.UpdateStatusOrder(db, lg, orderID, status)
	if err != nil {
		lg.Info("order AddOrder and add to channel", zap.String("about", ""))
		return err
	}
	return err
}
