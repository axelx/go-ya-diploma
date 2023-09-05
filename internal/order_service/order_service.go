package order_service

import (
	"fmt"
	"github.com/axelx/go-ya-diploma/internal/models"
	"github.com/axelx/go-ya-diploma/internal/pg"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"math"
	"math/rand"
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
	DB         *sqlx.DB
	LG         *zap.Logger
}

func (o Order) SearchIDs(orderNum string) (int, string) {

	ord, err := pg.FindOrder(o.DB, o.LG, orderNum)
	if err != nil {
		o.LG.Info("order SearchOne", zap.String("err", err.Error()))
		return 0, ""
	}
	return ord.UserID, ord.Number
}

func (o Order) SearchMany(userID int) ([]models.Order, error) {
	os, err := pg.FindOrders(o.DB, o.LG, userID)
	if err != nil {
		o.LG.Info("order FindOrders", zap.String("err", err.Error()))
	}

	for i, o := range os {
		if o.Accrual != 0 {
			os[i].Accrual = o.Accrual / 100
		}
		if o.Withdrawn > 0 {
			os[i].Withdrawn = o.Withdrawn / 100
		}
	}

	return os, nil
}

func (o Order) Create(userID int, orderID string, withdrawn float64, chAdd chan string) error {

	err := pg.AddOrder(o.DB, o.LG, userID, orderID, withdrawn)
	if err == nil {
		o.LG.Info("order AddOrder and add to channel", zap.String("orderID", orderID))
		chAdd <- orderID
	}
	return err
}

func (o Order) LunaCheck(order string) bool {
	sum := 0
	ord := 0
	if len(order) == 1 {
		return false
	}
	for i := len(order); i > 0; i-- {
		num, err := strconv.Atoi(string(order[i-1]))
		if err != nil {
			o.LG.Info("order_service LunaCheck ошибка", zap.String("about", ""))
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

func (o Order) FindOrder(orderID string) (models.Order, error) {
	or, err := pg.FindOrder(o.DB, o.LG, orderID)
	if err != nil {
		o.LG.Info("order FindOrder", zap.String("err", err.Error()))
		return or, err
	}
	return or, err
}

func FindWithdrawalsOrders(db *sqlx.DB, lg *zap.Logger, userID int) ([]models.OrderWithdrawal, error) {
	os, err := pg.FindOrders(db, lg, userID)
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
	}

	return res, nil
}

func UpdateStatus(db *sqlx.DB, lg *zap.Logger, orderID, status string, accrual float64) error {
	accrual = accrual * 100
	accrualInt := int(math.Round(accrual))
	err := pg.UpdateStatusOrder(db, lg, orderID, status, accrualInt)
	if err != nil {
		lg.Info("order AddOrder and add to channel", zap.String("about", ""))
		return err
	}
	return err
}

func GenerateLunaNumber(length int) string {
	//! с генерацией четных чисел проблема. (нужно начинать подсчёт с первого числа и добавлять в конец контрольное число.
	rand.Seed(time.Now().UnixNano()) // Инициализация генератора случайных чисел

	var num []int
	if length%2 == 0 {
		num = make([]int, length-1)
	} else {
		num = make([]int, length-1)
	}

	for i := 0; i < length-1; i++ {
		if i == 0 {
			num[i] = rand.Intn(9) + 1
		} else {
			num[i] = rand.Intn(10)
		}
	}

	sum := 0
	ord := 0
	for i := len(num); i > 0; i-- {
		digit := num[i-1]
		if (ord)%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		ord += 1
	}
	lastDigit := (10 - sum%10) % 10

	numStr := ""
	for _, digit := range num {
		numStr += fmt.Sprintf("%d", digit)
	}
	return fmt.Sprintf("%d", lastDigit) + numStr
}
