package orders

import (
	"github.com/axelx/go-ya-diploma/internal/core"
	"github.com/axelx/go-ya-diploma/internal/models"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"strconv"
)

func LunaCheck(order string, lg *zap.Logger) bool {
	a := []rune(order)

	sum := 0
	for i, r := range a {

		num, err := strconv.Atoi(string(r))
		if err != nil {
			return false
		}
		digit := num
		if (i+1)%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}

	if sum%10 != 0 {
		return false
	}
	return true
}

func FindOrder(db *sqlx.DB, lg *zap.Logger, order_id string) (models.Order, error) {
	o, err := core.FindOrder(db, lg, order_id)
	if err != nil {
		lg.Error("order FindOrder", zap.String("err", err.Error()))
	}
	return o, err
}

func FindOrders(db *sqlx.DB, lg *zap.Logger, user_id int) ([]models.Order, error) {
	os, err := core.FindOrders(db, lg, user_id)
	if err != nil {
		lg.Error("order FindOrders", zap.String("err", err.Error()))
	}

	return os, nil
}

func AddOrder(db *sqlx.DB, lg *zap.Logger, user_id int, order_id string, withdrawn float64) error {
	err := core.AddOrder(db, lg, user_id, order_id, withdrawn)
	return err
}
