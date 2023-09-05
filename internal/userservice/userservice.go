package userservice

import (
	"github.com/axelx/go-ya-diploma/internal/models"
	"github.com/axelx/go-ya-diploma/internal/pg"
	"github.com/axelx/go-ya-diploma/internal/utils"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type User struct {
	ID       int    `json:"id"`
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
	DB       *sqlx.DB
	LG       *zap.Logger
}

func (u User) SearchOne(login string) (int, string) {
	usrID, log := pg.FindUserByLogin(u.DB, u.LG, login)
	return usrID, log
}
func (u User) SearchMany(s string) ([]int, []string) {
	return []int{5}, []string{"user_" + s}
}

func (u User) Create(login, password string) error {
	err := pg.CreateNewUser(u.DB, u.LG, login, password)
	return err
}

func (u User) AuthUser(login, password string) (http.Cookie, bool) {
	usr := pg.AuthUser(u.DB, u.LG, login, password)

	if usr.Login == "" {
		return http.Cookie{}, false
	}

	cookie := http.Cookie{
		Name:    "auth",
		Value:   strconv.Itoa(usr.ID),
		Expires: time.Now().Add(time.Hour * 1),
		Path:    "/",
	}

	return cookie, true
}

func (u User) Balance(userID int) (models.Balance, error) {
	os, err := pg.FindOrders(u.DB, u.LG, userID)
	if err != nil {
		u.LG.Info("user Balance", zap.String("err", err.Error()))
	}

	b := models.Balance{}

	if len(os) > 0 {
		for _, o := range os {
			b.Current += float64(o.Accrual)
			b.Withdrawn += float64(o.Withdrawn)
		}
	}

	b.Current = (b.Current - b.Withdrawn) / 100
	b.Withdrawn = b.Withdrawn / 100

	return b, err
}

func (u User) GetIDviaCookie(cookiesVal string) int {
	userID := utils.StrToInt(cookiesVal)

	return userID
}
