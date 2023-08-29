package orders

import (
	"fmt"
	"github.com/axelx/go-ya-diploma/internal/core"
	"github.com/axelx/go-ya-diploma/internal/models"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"math"
	"strconv"
	"time"
)

type Order struct {
	ID         int        `json:"id,omitempty"`
	Number     string     `json:"order,omitempty"`
	Accrual    int        `json:"accrual"`
	Withdrawn  int        `json:"sum"`
	Status     string     `json:"status,omitempty"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`
	UserID     int        `json:"user_id,omitempty"`
}

func (o Order) SearchOne(db *sqlx.DB, lg *zap.Logger, orderNum string) (int, string) {

	ord, err := core.FindOrder(db, lg, orderNum)
	if err != nil {
		lg.Info("order SearchOne", zap.String("err", err.Error()))
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
	ord := 0
	for i := len(order); i > 0; i-- {
		num, err := strconv.Atoi(string(order[i-1]))
		if err != nil {
			lg.Info("orders LunaCheck ошибка", zap.String("about", ""))
			return false
		}
		digit := num
		if (ord)%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		ord += 1
	}

	return sum%10 == 0
}

func FindOrder(db *sqlx.DB, lg *zap.Logger, orderID string) (models.Order, error) {
	o, err := core.FindOrder(db, lg, orderID)
	if err != nil {
		lg.Info("order FindOrder", zap.String("err", err.Error()))
		return o, err
	}
	return o, err
}

func FindOrders(db *sqlx.DB, lg *zap.Logger, userID int, chAdd chan string) ([]models.Order, error) {
	os, err := core.FindOrders(db, lg, userID)
	if err != nil {
		lg.Info("order FindOrders", zap.String("err", err.Error()))
	}

	for i, o := range os {
		if o.Accrual != 0 {
			os[i].Accrual = o.Accrual / 100
		}
		if o.Withdrawn > 0 {
			os[i].Withdrawn = o.Withdrawn / 100
		}
		fmt.Println("----orders FindOrders():", o)
	}

	return os, nil
}

func FindWithdrawalsOrders(db *sqlx.DB, lg *zap.Logger, userID int) ([]models.OrderWithdrawal, error) {
	os, err := core.FindOrders(db, lg, userID)
	if err != nil {
		lg.Info("order FindWithdrawalsOrders", zap.String("err", err.Error()))
	}

	res := []models.OrderWithdrawal{}

	for _, o := range os {
		ow := models.OrderWithdrawal{}
		if o.Withdrawn > 0 {
			o.Withdrawn = o.Withdrawn / 100
			ow.Withdrawn = o.Withdrawn
			ow.Number = o.Number
			ow.UploadedAt = o.UploadedAt
			res = append(res, ow)
		}
		fmt.Println("----orders FindOrders():", o)
	}

	return res, nil
}

func AddOrder(db *sqlx.DB, lg *zap.Logger, userID int, orderID string, withdrawn float64, chAdd chan string) error {
	err := core.AddOrder(db, lg, userID, orderID, withdrawn)
	if err == nil {
		lg.Info("order AddOrder and add to channel", zap.String("about", ""))
		fmt.Println("----orders FindOrders(). добавляем в поток для начисления заказ. userIDЖ", userID, "orderID:", orderID)
		chAdd <- orderID
	}
	return err
}

func UpdateStatus(db *sqlx.DB, lg *zap.Logger, orderID, status string, accrual float64) error {
	accrual = accrual * 100
	accrualInt := int(math.Round(accrual))
	err := core.UpdateStatusOrder(db, lg, orderID, status, accrualInt)
	if err != nil {
		lg.Info("order AddOrder and add to channel", zap.String("about", ""))
		return err
	}
	return err
}
