package user

import (
	"github.com/axelx/go-ya-diploma/internal/core"
	"github.com/axelx/go-ya-diploma/internal/models"
	"github.com/axelx/go-ya-diploma/internal/utils"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func FindUserByLogin(db *sqlx.DB, lg *zap.Logger, login string) (models.User, error) {
	u, err := core.FindUserByLogin(db, lg, login)
	return u, err
}

func CreateNewUser(db *sqlx.DB, lg *zap.Logger, login, password string) error {
	err := core.CreateNewUser(db, lg, login, password)
	return err
}

func AuthUser(db *sqlx.DB, lg *zap.Logger, login, password string) (http.Cookie, bool) {
	u := core.AuthUser(db, lg, login, password)

	if u.Login == "" {
		return http.Cookie{}, false
	}

	cookie := http.Cookie{
		Name:    "auth",
		Value:   u.ID,
		Expires: time.Now().Add(time.Hour * 1),
		Path:    "/",
	}

	return cookie, true
}

func Balance(db *sqlx.DB, lg *zap.Logger, userID int) (models.Balance, error) {
	os, err := core.FindOrders(db, lg, userID)
	if err != nil {
		lg.Info("user Balance", zap.String("err", err.Error()))
	}

	b := models.Balance{}

	if len(os) > 0 {
		for _, o := range os {
			b.Current += float64(o.Accrual)
			b.Withdrawn += float64(o.Withdrawn)
		}
	}

	b.Current = b.Current / 100
	b.Withdrawn = b.Withdrawn / 100

	return b, err
}

func GetIDviaCookie(req *http.Request) int {
	cookies, _ := req.Cookie("auth")
	user_id := utils.StrToInt(cookies.Value)

	return user_id
}
