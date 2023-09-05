package pg

import (
	"context"
	"github.com/axelx/go-ya-diploma/internal/models"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func FindUserByLogin(db *sqlx.DB, lg *zap.Logger, login string) (int, string) {
	row := db.QueryRowContext(context.Background(), `SELECT id, login FROM users WHERE login = $1`, login)
	var v models.User
	err := row.Scan(&v.ID, &v.Login)
	if err != nil {
		lg.Info("Error FindUserByLogin : user not found", zap.String("about ERR", err.Error()), zap.String("login", login))
		return 0, ""
	}
	lg.Info("pg FindUserByLogin :", zap.String("user_id", v.Login))
	return v.ID, v.Login
}

func CreateNewUser(db *sqlx.DB, lg *zap.Logger, login, password string) error {
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO users (login, password) VALUES ($1, $2)`, login, password)
	if err != nil {
		lg.Info("Error CreateNewUser :", zap.String("about ERR", err.Error()))

		return err
	}
	return nil
}

func AuthUser(db *sqlx.DB, lg *zap.Logger, login, password string) models.User {
	row := db.QueryRowContext(context.Background(),
		` SELECT * FROM users WHERE login = $1 AND password = $2 `, login, password)

	var usr models.User
	err := row.Scan(&usr.ID, &usr.Login, &usr.Password)
	if err != nil {
		lg.Info("Error AuthUser :", zap.String("about ERR", err.Error()))
		return models.User{}
	}
	return usr
}

func Balance(db *sqlx.DB, lg *zap.Logger, userID int) ([]models.Order, error) {
	rows, err := db.QueryContext(context.Background(),
		` SELECT number,accrual,status,uploaded_at FROM orders WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	orders := []models.Order{}

	for rows.Next() {
		var o models.Order
		err = rows.Scan(&o.Number, &o.Accrual, &o.Status, &o.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return orders, nil
}
